package database

import (
	"context"
	"database/sql"
	"errors"
	"example/sensorHub/types"
	"fmt"
	"log/slog"
)

type MQTTSubscriptionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMQTTSubscriptionRepository(db *sql.DB, logger *slog.Logger) *MQTTSubscriptionRepository {
	return &MQTTSubscriptionRepository{db: db, logger: logger.With("component", "mqtt_subscription_repository")}
}

func (r *MQTTSubscriptionRepository) Add(ctx context.Context, sub types.MQTTSubscription) (int, error) {
	if sub.TopicPattern == "" {
		return 0, fmt.Errorf("topic pattern cannot be empty")
	}
	if sub.DriverType == "" {
		return 0, fmt.Errorf("driver type cannot be empty")
	}
	if sub.BrokerId <= 0 {
		return 0, fmt.Errorf("broker id must be positive")
	}
	query := `INSERT INTO mqtt_subscriptions (broker_id, topic_pattern, driver_type, enabled)
		VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, sub.BrokerId, sub.TopicPattern, sub.DriverType, sub.Enabled)
	if err != nil {
		return 0, fmt.Errorf("error adding MQTT subscription: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id for MQTT subscription: %w", err)
	}
	return int(id), nil
}

func (r *MQTTSubscriptionRepository) GetByID(ctx context.Context, id int) (*types.MQTTSubscription, error) {
	query := `SELECT id, broker_id, topic_pattern, driver_type, enabled, created_at, updated_at
		FROM mqtt_subscriptions WHERE id = ?`
	sub, err := scanSubscriptionRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying MQTT subscription by id: %w", err)
	}
	return &sub, nil
}

func (r *MQTTSubscriptionRepository) GetAll(ctx context.Context) ([]types.MQTTSubscription, error) {
	query := `SELECT id, broker_id, topic_pattern, driver_type, enabled, created_at, updated_at
		FROM mqtt_subscriptions ORDER BY broker_id, topic_pattern`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all MQTT subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []types.MQTTSubscription
	for rows.Next() {
		sub, err := scanSubscriptionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning MQTT subscription row: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over MQTT subscription rows: %w", err)
	}
	return subs, nil
}

func (r *MQTTSubscriptionRepository) GetByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	query := `SELECT id, broker_id, topic_pattern, driver_type, enabled, created_at, updated_at
		FROM mqtt_subscriptions WHERE broker_id = ? ORDER BY topic_pattern`
	rows, err := r.db.QueryContext(ctx, query, brokerID)
	if err != nil {
		return nil, fmt.Errorf("error querying MQTT subscriptions by broker id: %w", err)
	}
	defer rows.Close()

	var subs []types.MQTTSubscription
	for rows.Next() {
		sub, err := scanSubscriptionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning MQTT subscription row: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over MQTT subscription rows: %w", err)
	}
	return subs, nil
}

func (r *MQTTSubscriptionRepository) GetEnabledByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	query := `SELECT id, broker_id, topic_pattern, driver_type, enabled, created_at, updated_at
		FROM mqtt_subscriptions WHERE broker_id = ? AND enabled = 1 ORDER BY topic_pattern`
	rows, err := r.db.QueryContext(ctx, query, brokerID)
	if err != nil {
		return nil, fmt.Errorf("error querying enabled MQTT subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []types.MQTTSubscription
	for rows.Next() {
		sub, err := scanSubscriptionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning MQTT subscription row: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over MQTT subscription rows: %w", err)
	}
	return subs, nil
}

func (r *MQTTSubscriptionRepository) Update(ctx context.Context, sub types.MQTTSubscription) error {
	query := `UPDATE mqtt_subscriptions SET broker_id = ?, topic_pattern = ?, driver_type = ?,
		enabled = ?, updated_at = datetime('now') WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, sub.BrokerId, sub.TopicPattern, sub.DriverType, sub.Enabled, sub.Id)
	if err != nil {
		return fmt.Errorf("error updating MQTT subscription: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after MQTT subscription update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no MQTT subscription found with id %d", sub.Id)
	}
	return nil
}

func (r *MQTTSubscriptionRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM mqtt_subscriptions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting MQTT subscription: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after MQTT subscription delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no MQTT subscription found with id %d", id)
	}
	return nil
}

func scanSubscriptionRow(row scannable) (types.MQTTSubscription, error) {
	var s types.MQTTSubscription
	var createdAt, updatedAt NullSQLiteTime
	err := row.Scan(
		&s.Id, &s.BrokerId, &s.TopicPattern, &s.DriverType,
		&s.Enabled, &createdAt, &updatedAt,
	)
	if err != nil {
		return s, err
	}
	if createdAt.Valid {
		s.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		s.UpdatedAt = updatedAt.Time
	}
	return s, nil
}
