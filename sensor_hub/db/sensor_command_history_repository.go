package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type SensorCommandHistoryRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSensorCommandHistoryRepository(db *sql.DB, logger *slog.Logger) *SensorCommandHistoryRepository {
	return &SensorCommandHistoryRepository{db: db, logger: logger.With("component", "sensor_command_history_repository")}
}

func (r *SensorCommandHistoryRepository) AddSentCommand(ctx context.Context, sensorID int, userID *int, property string, value string, mqttTopic string, mqttPayload string, timeoutSeconds int, sentAt time.Time) (int, error) {
	var userIDValue any
	if userID != nil {
		userIDValue = *userID
	}

	query := `INSERT INTO sensor_command_history
		(sensor_id, user_id, property, value, status, mqtt_topic, mqtt_payload, timeout_seconds, sent_at)
		VALUES (?, ?, ?, ?, 'sent', ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, sensorID, userIDValue, property, value, mqttTopic, mqttPayload, timeoutSeconds, sentAt)
	if err != nil {
		return 0, fmt.Errorf("error inserting sensor command history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error reading sensor command history insert id: %w", err)
	}

	return int(id), nil
}

func (r *SensorCommandHistoryRepository) HasPendingCommand(ctx context.Context, sensorID int, property string) (bool, error) {
	query := `SELECT COUNT(1) FROM sensor_command_history WHERE sensor_id = ? AND property = ? AND status = 'sent'`
	var count int
	if err := r.db.QueryRowContext(ctx, query, sensorID, property).Scan(&count); err != nil {
		return false, fmt.Errorf("error querying pending sensor commands: %w", err)
	}
	return count > 0, nil
}
