package database

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSensorCommandHistoryRepository_AddSentCommand_Success(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSensorCommandHistoryRepository(db, slog.Default())
	userID := 99
	sentAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO sensor_command_history").
		WithArgs(7, &userID, "state", "ON", "zigbee2mqtt/office-plug/set", `{"state":"ON"}`, 10, sentAt).
		WillReturnResult(sqlmock.NewResult(42, 1))

	id, err := repo.AddSentCommand(context.Background(), 7, &userID, "state", "ON", "zigbee2mqtt/office-plug/set", `{"state":"ON"}`, 10, sentAt)
	assert.NoError(t, err)
	assert.Equal(t, 42, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSensorCommandHistoryRepository_HasPendingCommand_ReturnsTrue(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSensorCommandHistoryRepository(db, slog.Default())

	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM sensor_command_history WHERE sensor_id = \\? AND property = \\? AND status = 'sent'").
		WithArgs(7, "state").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	hasPending, err := repo.HasPendingCommand(context.Background(), 7, "state")
	assert.NoError(t, err)
	assert.True(t, hasPending)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSensorCommandHistoryRepository_AddSentCommand_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSensorCommandHistoryRepository(db, slog.Default())

	mock.ExpectExec("INSERT INTO sensor_command_history").
		WithArgs(7, nil, "state", "ON", "zigbee2mqtt/office-plug/set", `{"state":"ON"}`, 10, sqlmock.AnyArg()).
		WillReturnError(errors.New("write failed"))

	_, err := repo.AddSentCommand(context.Background(), 7, nil, "state", "ON", "zigbee2mqtt/office-plug/set", `{"state":"ON"}`, 10, time.Now().UTC())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error inserting sensor command history")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSensorCommandHistoryRepository_MarkAcknowledged_Success(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSensorCommandHistoryRepository(db, slog.Default())
	acknowledgedAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	mock.ExpectExec("UPDATE sensor_command_history").
		WithArgs(acknowledgedAt, "OFF", 42).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updated, err := repo.MarkAcknowledged(context.Background(), 42, "OFF", acknowledgedAt)
	assert.NoError(t, err)
	assert.True(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSensorCommandHistoryRepository_ListPendingCommands_ReturnsRows(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSensorCommandHistoryRepository(db, slog.Default())
	sentAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, sensor_id, property, value, status, timeout_seconds, sent_at, acknowledged_at, acknowledged_value").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sensor_id", "property", "value", "status", "timeout_seconds", "sent_at", "acknowledged_at", "acknowledged_value",
		}).AddRow(42, 7, "state", "ON", "sent", 10, sentAt, nil, nil))

	commands, err := repo.ListPendingCommands(context.Background())
	assert.NoError(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, 42, commands[0].ID)
	assert.Equal(t, 7, commands[0].SensorID)
	assert.Equal(t, "state", commands[0].Property)
	assert.Equal(t, "ON", commands[0].Value)
	assert.Equal(t, "sent", commands[0].Status)
	assert.Equal(t, 10, commands[0].TimeoutSeconds)
	assert.Equal(t, sentAt, commands[0].SentAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}
