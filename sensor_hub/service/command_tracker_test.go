package service

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	database "example/sensorHub/db"
	gen "example/sensorHub/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAckOnReading_MarksAcknowledged(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	repo := newFakeCommandTrackerRepository(database.PendingCommandRecord{
		ID:             42,
		SensorID:       7,
		Property:       "state",
		Value:          "ON",
		TimeoutSeconds: 10,
		SentAt:         now.Add(-2 * time.Second),
	})
	broadcaster := &fakeCommandStatusBroadcaster{}
	tracker := NewCommandTracker(repo, broadcaster, slog.Default())
	tracker.now = func() time.Time { return now }
	defer tracker.Close()

	tracker.Track(context.Background(), repo.mustGet(42))
	tracker.ObserveReadings(context.Background(), 7, []gen.Reading{{
		MeasurementType: "state",
		TextState:       ptrString("OFF"),
	}})

	command := repo.mustGet(42)
	require.Equal(t, commandStatusAcknowledged, command.Status)
	require.NotNil(t, command.AcknowledgedAt)
	require.Equal(t, now, *command.AcknowledgedAt)
	require.NotNil(t, command.AcknowledgedValue)
	assert.Equal(t, "OFF", *command.AcknowledgedValue)

	require.Len(t, broadcaster.messages, 1)
	assert.Equal(t, "command_status", broadcaster.messages[0].Type)
	assert.Equal(t, commandStatusAcknowledged, broadcaster.messages[0].Status)
	assert.Equal(t, 42, broadcaster.messages[0].ID)
}

func TestAckTimeout_MarksTimedOut(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	repo := newFakeCommandTrackerRepository(database.PendingCommandRecord{
		ID:             43,
		SensorID:       7,
		Property:       "state",
		Value:          "ON",
		TimeoutSeconds: 10,
		SentAt:         now,
	})
	broadcaster := &fakeCommandStatusBroadcaster{}
	tracker := NewCommandTracker(repo, broadcaster, slog.Default())
	tracker.now = func() time.Time { return now }
	tracker.schedule = func(_ time.Duration, fn func()) func() {
		go fn()
		return func() {}
	}

	tracker.Track(context.Background(), repo.mustGet(43))

	require.Eventually(t, func() bool {
		command := repo.mustGet(43)
		return command.Status == commandStatusTimedOut && len(broadcaster.messages) == 1
	}, time.Second, 10*time.Millisecond)

	assert.Equal(t, "command_status", broadcaster.messages[0].Type)
	assert.Equal(t, commandStatusTimedOut, broadcaster.messages[0].Status)
	assert.Equal(t, 43, broadcaster.messages[0].ID)
}

func TestAckOnReading_MatchesPropertyOnly(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	repo := newFakeCommandTrackerRepository(database.PendingCommandRecord{
		ID:             44,
		SensorID:       7,
		Property:       "state",
		Value:          "ON",
		TimeoutSeconds: 10,
		SentAt:         now,
	})
	broadcaster := &fakeCommandStatusBroadcaster{}
	tracker := NewCommandTracker(repo, broadcaster, slog.Default())
	tracker.now = func() time.Time { return now }
	defer tracker.Close()

	tracker.Track(context.Background(), repo.mustGet(44))
	tracker.ObserveReadings(context.Background(), 7, []gen.Reading{{
		MeasurementType: "state",
		TextState:       ptrString("OFF"),
	}})

	command := repo.mustGet(44)
	require.Equal(t, commandStatusAcknowledged, command.Status)
	require.NotNil(t, command.AcknowledgedValue)
	assert.Equal(t, "OFF", *command.AcknowledgedValue)
}

func TestRecoverPending_TimesOutExpiredCommandsAndTracksRemainingOnes(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	repo := newFakeCommandTrackerRepository(
		database.PendingCommandRecord{
			ID:             45,
			SensorID:       7,
			Property:       "state",
			Value:          "ON",
			TimeoutSeconds: 1,
			SentAt:         now.Add(-2 * time.Second),
		},
		database.PendingCommandRecord{
			ID:             46,
			SensorID:       8,
			Property:       "state",
			Value:          "OFF",
			TimeoutSeconds: 10,
			SentAt:         now,
		},
	)
	broadcaster := &fakeCommandStatusBroadcaster{}
	tracker := NewCommandTracker(repo, broadcaster, slog.Default())
	tracker.now = func() time.Time { return now }
	tracker.schedule = func(_ time.Duration, _ func()) func() { return func() {} }
	defer tracker.Close()

	require.NoError(t, tracker.RecoverPending(context.Background()))

	expired := repo.mustGet(45)
	assert.Equal(t, commandStatusTimedOut, expired.Status)

	tracker.ObserveReadings(context.Background(), 8, []gen.Reading{{
		MeasurementType: "state",
		TextState:       ptrString("OFF"),
	}})

	recovered := repo.mustGet(46)
	assert.Equal(t, commandStatusAcknowledged, recovered.Status)
	require.Len(t, broadcaster.messages, 2)
	assert.Equal(t, commandStatusTimedOut, broadcaster.messages[0].Status)
	assert.Equal(t, commandStatusAcknowledged, broadcaster.messages[1].Status)
}

type fakeCommandTrackerRepository struct {
	mu       sync.Mutex
	commands map[int]database.PendingCommandRecord
}

func newFakeCommandTrackerRepository(commands ...database.PendingCommandRecord) *fakeCommandTrackerRepository {
	repo := &fakeCommandTrackerRepository{
		commands: make(map[int]database.PendingCommandRecord, len(commands)),
	}
	for _, command := range commands {
		command.Status = commandStatusSent
		repo.commands[command.ID] = command
	}
	return repo
}

func (r *fakeCommandTrackerRepository) MarkAcknowledged(_ context.Context, id int, acknowledgedValue string, acknowledgedAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	command, ok := r.commands[id]
	if !ok || command.Status != commandStatusSent {
		return false, nil
	}

	command.Status = commandStatusAcknowledged
	command.AcknowledgedAt = &acknowledgedAt
	command.AcknowledgedValue = &acknowledgedValue
	r.commands[id] = command
	return true, nil
}

func (r *fakeCommandTrackerRepository) MarkTimedOut(_ context.Context, id int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	command, ok := r.commands[id]
	if !ok || command.Status != commandStatusSent {
		return false, nil
	}

	command.Status = commandStatusTimedOut
	r.commands[id] = command
	return true, nil
}

func (r *fakeCommandTrackerRepository) MarkFailed(_ context.Context, id int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	command, ok := r.commands[id]
	if !ok || command.Status != commandStatusSent {
		return false, nil
	}

	command.Status = commandStatusFailed
	r.commands[id] = command
	return true, nil
}

func (r *fakeCommandTrackerRepository) ListPendingCommands(_ context.Context) ([]database.PendingCommandRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	commands := make([]database.PendingCommandRecord, 0, len(r.commands))
	for _, command := range r.commands {
		if command.Status == commandStatusSent {
			commands = append(commands, command)
		}
	}
	return commands, nil
}

func (r *fakeCommandTrackerRepository) mustGet(id int) database.PendingCommandRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.commands[id]
}

type fakeCommandStatusBroadcaster struct {
	messages []CommandStatusMessage
}

func (b *fakeCommandStatusBroadcaster) BroadcastCommandStatus(message CommandStatusMessage) {
	b.messages = append(b.messages, message)
}

func ptrString(value string) *string {
	return &value
}
