package database

import (
	"context"
	"database/sql"
	"errors"
	"example/sensorHub/types"
	"fmt"
	"log/slog"
)

type MQTTBrokerRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMQTTBrokerRepository(db *sql.DB, logger *slog.Logger) *MQTTBrokerRepository {
	return &MQTTBrokerRepository{db: db, logger: logger.With("component", "mqtt_broker_repository")}
}

func (r *MQTTBrokerRepository) Add(ctx context.Context, broker types.MQTTBroker) (int, error) {
	if broker.Name == "" {
		return 0, fmt.Errorf("broker name cannot be empty")
	}
	if broker.Host == "" {
		return 0, fmt.Errorf("broker host cannot be empty")
	}
	if broker.Port <= 0 {
		broker.Port = 1883
	}
	if broker.Type == "" {
		broker.Type = "external"
	}
	query := `INSERT INTO mqtt_brokers (name, type, host, port, username, password, client_id,
		ca_cert_path, client_cert_path, client_key_path, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		broker.Name, broker.Type, broker.Host, broker.Port,
		nullString(broker.Username), nullString(broker.Password), nullString(broker.ClientId),
		nullString(broker.CACertPath), nullString(broker.ClientCertPath), nullString(broker.ClientKeyPath),
		broker.Enabled,
	)
	if err != nil {
		return 0, fmt.Errorf("error adding MQTT broker: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id for MQTT broker: %w", err)
	}
	return int(id), nil
}

func (r *MQTTBrokerRepository) GetByID(ctx context.Context, id int) (*types.MQTTBroker, error) {
	query := `SELECT id, name, type, host, port, username, password, client_id,
		ca_cert_path, client_cert_path, client_key_path, enabled, created_at, updated_at
		FROM mqtt_brokers WHERE id = ?`
	broker, err := scanBrokerRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying MQTT broker by id: %w", err)
	}
	return &broker, nil
}

func (r *MQTTBrokerRepository) GetByName(ctx context.Context, name string) (*types.MQTTBroker, error) {
	query := `SELECT id, name, type, host, port, username, password, client_id,
		ca_cert_path, client_cert_path, client_key_path, enabled, created_at, updated_at
		FROM mqtt_brokers WHERE LOWER(name) = LOWER(?)`
	broker, err := scanBrokerRow(r.db.QueryRowContext(ctx, query, name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying MQTT broker by name: %w", err)
	}
	return &broker, nil
}

func (r *MQTTBrokerRepository) GetAll(ctx context.Context) ([]types.MQTTBroker, error) {
	query := `SELECT id, name, type, host, port, username, password, client_id,
		ca_cert_path, client_cert_path, client_key_path, enabled, created_at, updated_at
		FROM mqtt_brokers ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all MQTT brokers: %w", err)
	}
	defer rows.Close()

	var brokers []types.MQTTBroker
	for rows.Next() {
		broker, err := scanBrokerRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning MQTT broker row: %w", err)
		}
		brokers = append(brokers, broker)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over MQTT broker rows: %w", err)
	}
	return brokers, nil
}

func (r *MQTTBrokerRepository) Update(ctx context.Context, broker types.MQTTBroker) error {
	query := `UPDATE mqtt_brokers SET name = ?, type = ?, host = ?, port = ?,
		username = ?, password = ?, client_id = ?,
		ca_cert_path = ?, client_cert_path = ?, client_key_path = ?,
		enabled = ?, updated_at = datetime('now') WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query,
		broker.Name, broker.Type, broker.Host, broker.Port,
		nullString(broker.Username), nullString(broker.Password), nullString(broker.ClientId),
		nullString(broker.CACertPath), nullString(broker.ClientCertPath), nullString(broker.ClientKeyPath),
		broker.Enabled, broker.Id,
	)
	if err != nil {
		return fmt.Errorf("error updating MQTT broker: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after MQTT broker update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no MQTT broker found with id %d", broker.Id)
	}
	return nil
}

func (r *MQTTBrokerRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM mqtt_brokers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting MQTT broker: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after MQTT broker delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no MQTT broker found with id %d", id)
	}
	return nil
}

func (r *MQTTBrokerRepository) GetEnabled(ctx context.Context) ([]types.MQTTBroker, error) {
	query := `SELECT id, name, type, host, port, username, password, client_id,
		ca_cert_path, client_cert_path, client_key_path, enabled, created_at, updated_at
		FROM mqtt_brokers WHERE enabled = 1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying enabled MQTT brokers: %w", err)
	}
	defer rows.Close()

	var brokers []types.MQTTBroker
	for rows.Next() {
		broker, err := scanBrokerRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning MQTT broker row: %w", err)
		}
		brokers = append(brokers, broker)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over MQTT broker rows: %w", err)
	}
	return brokers, nil
}

// nullString converts an empty string to a sql.NullString for nullable TEXT columns.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func scanBrokerRow(row scannable) (types.MQTTBroker, error) {
	var b types.MQTTBroker
	var username, password, clientId sql.NullString
	var caCert, clientCert, clientKey sql.NullString
	var createdAt, updatedAt NullSQLiteTime
	err := row.Scan(
		&b.Id, &b.Name, &b.Type, &b.Host, &b.Port,
		&username, &password, &clientId,
		&caCert, &clientCert, &clientKey,
		&b.Enabled, &createdAt, &updatedAt,
	)
	if err != nil {
		return b, err
	}
	b.Username = username.String
	b.Password = password.String
	b.ClientId = clientId.String
	b.CACertPath = caCert.String
	b.ClientCertPath = clientCert.String
	b.ClientKeyPath = clientKey.String
	if createdAt.Valid {
		b.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		b.UpdatedAt = updatedAt.Time
	}
	return b, nil
}
