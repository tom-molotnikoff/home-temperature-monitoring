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
