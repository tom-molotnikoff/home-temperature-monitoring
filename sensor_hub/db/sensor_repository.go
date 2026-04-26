package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	gen "example/sensorHub/gen"
	"fmt"
	"log/slog"
	"time"
)

type SensorRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSensorRepository(db *sql.DB, logger *slog.Logger) *SensorRepository {
	return &SensorRepository{db: db, logger: logger.With("component", "sensor_repository")}
}

func (s *SensorRepository) SensorExists(ctx context.Context, name string) (bool, error) {
	query := "SELECT COUNT(1) FROM sensors WHERE LOWER(name) = LOWER(?)"
	var count int
	err := s.db.QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking if sensor exists: %w", err)
	}
	return count > 0, nil
}

func (s *SensorRepository) SetEnabledSensorByName(ctx context.Context, name string, enabled bool) error {
	query := "UPDATE sensors SET enabled = ?, health_status = ? WHERE LOWER(name) = LOWER(?)"
	if !enabled {
		go func(name string, status gen.SensorHealthStatus) {
			sensorId, err := s.GetSensorIdByName(context.Background(), name)
			if err != nil {
				s.logger.Error("failed to get sensor id for health history insert", "error", err)
				return
			}
			if sensorId <= 0 {
				s.logger.Warn("skipping sensor health history insert: invalid sensor id", "sensor_id", sensorId)
				return
			}
			insertQuery := fmt.Sprintf("INSERT INTO %s (sensor_id, health_status) VALUES (?, ?)", TableSensorHealthHistory)
			if _, err := s.db.ExecContext(context.Background(), insertQuery, sensorId, status); err != nil {
				s.logger.Error("failed to insert sensor health history", "sensor_id", sensorId, "error", err)
			}
		}(name, gen.Unknown)
		query = "UPDATE sensors SET enabled = ?, health_status = ?, health_reason = 'unknown' WHERE LOWER(name) = LOWER(?)"
	}
	result, err := s.db.ExecContext(ctx, query, enabled, gen.Unknown, name)
	if err != nil {
		return fmt.Errorf("error updating sensor enabled status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no changes were made to sensor %s", name)
	}
	return nil
}

func (s *SensorRepository) GetSensorIdByName(ctx context.Context, sensorName string) (int, error) {
	query := "SELECT id FROM sensors WHERE LOWER(name) = LOWER(?)"
	var sensorID int
	err := s.db.QueryRowContext(ctx, query, sensorName).Scan(&sensorID)
	if err != nil {
		return 0, fmt.Errorf("could not find sensor id for name %s: %w", sensorName, err)
	}
	return sensorID, nil
}

func (s *SensorRepository) DeleteHealthHistoryOlderThan(ctx context.Context, cutoffDate time.Time) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE recorded_at < ?", TableSensorHealthHistory)
	_, err := s.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("error deleting old sensor health history: %w", err)
	}
	return nil
}

func (s *SensorRepository) GetSensorHealthHistoryById(ctx context.Context, sensorId int, limit int) ([]gen.SensorHealthHistory, error) {
	query := fmt.Sprintf("SELECT id, sensor_id, health_status, recorded_at FROM %s WHERE sensor_id = ? ORDER BY recorded_at DESC LIMIT ?", TableSensorHealthHistory)
	rows, err := s.db.QueryContext(ctx, query, sensorId, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying sensor health history: %w", err)
	}
	defer rows.Close()

	var history []gen.SensorHealthHistory
	for rows.Next() {
		var record gen.SensorHealthHistory
		var recordedAt SQLiteTime
		if err := rows.Scan(&record.Id, &record.SensorId, &record.HealthStatus, &recordedAt); err != nil {
			return nil, fmt.Errorf("error scanning sensor health history row: %w", err)
		}
		record.RecordedAt = recordedAt.Time
		history = append(history, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor health history rows: %w", err)
	}
	return history, nil
}

func (s *SensorRepository) DeleteSensorByName(ctx context.Context, name string) error {
	sensorId, err := s.GetSensorIdByName(ctx, name)
	if err != nil {
		return fmt.Errorf("error retrieving sensor ID for deletion: %w", err)
	}

	/*
	 TODO: transaction is good but purge should be its own service
	 so as to not hold up the API call whilst potentially deleting
	 a lot of data. eg: schedule a purge and let the other service
	 handle it asynchronously.
	*/

	txn, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			txn.Rollback()
			panic(p)
		} else if err != nil {
			txn.Rollback()
		} else {
			err = txn.Commit()
		}
	}()

	purgeQuery := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ?", TableReadings)
	_, err = txn.Exec(purgeQuery, sensorId)
	if err != nil {
		return fmt.Errorf("error purging readings for sensor ID %d: %w", sensorId, err)
	}
	sensorMtPurgeQuery := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ?", TableSensorMeasurementTypes)
	_, err = txn.Exec(sensorMtPurgeQuery, sensorId)
	if err != nil {
		return fmt.Errorf("error purging sensor measurement types for sensor ID %d: %w", sensorId, err)
	}
	healthHistoryPurgeQuery := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ?", TableSensorHealthHistory)
	_, err = txn.Exec(healthHistoryPurgeQuery, sensorId)
	if err != nil {
		return fmt.Errorf("error purging sensor health history for sensor ID %d: %w", sensorId, err)
	}

	query := "DELETE FROM sensors WHERE LOWER(name) = LOWER(?)"
	result, err := txn.Exec(query, name)
	if err != nil {
		return fmt.Errorf("error deleting sensor: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no sensor found with name %s to delete", name)
	}

	return nil
}

func (s *SensorRepository) GetSensorsByDriver(ctx context.Context, sensorDriver string) ([]gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE LOWER(sensor_driver) = LOWER(?)"
	rows, err := s.db.QueryContext(ctx, query, sensorDriver)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors by driver: %w", err)
	}
	defer rows.Close()

	var sensors []gen.Sensor
	for rows.Next() {
		sensor, err := scanSensorRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

func (s *SensorRepository) UpdateSensorById(ctx context.Context, sensor gen.Sensor, retentionHoursPresent bool) error {
	configJSON, err := json.Marshal(sensor.Config)
	if err != nil {
		return fmt.Errorf("error marshalling sensor config: %w", err)
	}

	var result sql.Result
	if retentionHoursPresent {
		// retention_hours was explicitly provided (even if null — meaning "clear it").
		query := "UPDATE sensors SET name = ?, sensor_driver = ?, config = ?, retention_hours = ? WHERE id = ?"
		result, err = s.db.ExecContext(ctx, query, sensor.Name, sensor.SensorDriver, string(configJSON), sensor.RetentionHours, sensor.Id)
	} else {
		query := "UPDATE sensors SET name = ?, sensor_driver = ?, config = ? WHERE id = ?"
		result, err = s.db.ExecContext(ctx, query, sensor.Name, sensor.SensorDriver, string(configJSON), sensor.Id)
	}
	if err != nil {
		return fmt.Errorf("error updating sensor: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no changes were made to sensor %s", sensor.Name)
	}
	return nil
}

func (s *SensorRepository) AddSensor(ctx context.Context, sensor gen.Sensor) error {
	if sensor.Name == "" || sensor.SensorDriver == "" {
		return fmt.Errorf("sensor name and sensor driver cannot be empty")
	}
	if sensor.Config == nil {
		sensor.Config = make(map[string]string)
	}

	configJSON, err := json.Marshal(sensor.Config)
	if err != nil {
		return fmt.Errorf("error marshalling sensor config: %w", err)
	}

	query := "INSERT INTO sensors (name, external_id, sensor_driver, config, health_reason, enabled, status) VALUES (?, ?, ?, ?, 'unknown', ?, ?)"
	status := sensor.Status
	if status == "" {
		status = gen.SensorStatusActive
	}
	_, err = s.db.ExecContext(ctx, query, sensor.Name, sensor.ExternalId, sensor.SensorDriver, string(configJSON), true, status)
	if err != nil {
		return fmt.Errorf("error adding new sensor: %w", err)
	}
	return nil
}

func (s *SensorRepository) GetSensorById(ctx context.Context, id int) (*gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE id = ?"
	sensor, err := scanSensorRow(s.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no sensor found with id %d", id)
		}
		return nil, fmt.Errorf("error querying sensor by id: %w", err)
	}
	return &sensor, nil
}

func (s *SensorRepository) GetSensorByName(ctx context.Context, name string) (*gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE LOWER(name) = LOWER(?)"
	sensor, err := scanSensorRow(s.db.QueryRowContext(ctx, query, name))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no sensor found with name %s", name)
		}
		return nil, fmt.Errorf("error querying sensor by name: %w", err)
	}
	return &sensor, nil
}

func (s *SensorRepository) GetSensorByExternalId(ctx context.Context, externalId string) (*gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE LOWER(external_id) = LOWER(?)"
	sensor, err := scanSensorRow(s.db.QueryRowContext(ctx, query, externalId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no sensor found with external_id %s: %w", externalId, err)
		}
		return nil, fmt.Errorf("error querying sensor by external_id: %w", err)
	}
	return &sensor, nil
}

func (s *SensorRepository) SensorExistsByExternalId(ctx context.Context, externalId string) (bool, error) {
	query := "SELECT COUNT(1) FROM sensors WHERE LOWER(external_id) = LOWER(?)"
	var count int
	err := s.db.QueryRowContext(ctx, query, externalId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking if sensor exists by external_id: %w", err)
	}
	return count > 0, nil
}

func (s *SensorRepository) GetAllSensors(ctx context.Context) ([]gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all sensors: %w", err)
	}
	defer rows.Close()

	var sensors []gen.Sensor
	for rows.Next() {
		sensor, err := scanSensorRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

func (s *SensorRepository) UpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus gen.SensorHealthStatus, healthReason string) error {
	query := "UPDATE sensors SET health_status = ?, health_reason = ? WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, healthStatus, healthReason, sensorId)
	if err != nil {
		return fmt.Errorf("error updating sensor health status: %w", err)
	}

	go func(id int, status gen.SensorHealthStatus) {
		if id <= 0 {
			s.logger.Warn("skipping sensor health history insert: invalid sensor id", "sensor_id", id)
			return
		}
		insertQuery := fmt.Sprintf("INSERT INTO %s (sensor_id, health_status) VALUES (?, ?)", TableSensorHealthHistory)
		if _, err := s.db.ExecContext(context.Background(), insertQuery, id, status); err != nil {
			s.logger.Error("failed to insert sensor health history", "sensor_id", id, "error", err)
		}
	}(sensorId, healthStatus)

	return nil
}

// TODO - implement methods for getting sensor health over time for reporting - see V5__sensor_health_history.sql

// scannable is satisfied by both *sql.Row and *sql.Rows.
type scannable interface {
	Scan(dest ...any) error
}

// scanSensorRow scans a sensor row (columns: id, name, external_id, sensor_driver, config,
// health_status, health_reason, enabled, status, retention_hours) and unmarshals the JSON
// config column into the Config map.
func scanSensorRow(row scannable) (gen.Sensor, error) {
	var s gen.Sensor
	var configJSON string
	var externalId sql.NullString
	var retentionHours sql.NullInt64
	err := row.Scan(&s.Id, &s.Name, &externalId, &s.SensorDriver, &configJSON, &s.HealthStatus, &s.HealthReason, &s.Enabled, &s.Status, &retentionHours)
	if err != nil {
		return s, err
	}
	if externalId.Valid {
		s.ExternalId = &externalId.String
	}
	if retentionHours.Valid {
		v := int(retentionHours.Int64)
		s.RetentionHours = &v
	}
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &s.Config); err != nil {
			return s, fmt.Errorf("failed to unmarshal sensor config: %w", err)
		}
	}
	if s.Config == nil {
		s.Config = make(map[string]string)
	}
	return s, nil
}

func (sr *SensorRepository) GetSensorsByStatus(ctx context.Context, status string) ([]gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE LOWER(status) = LOWER(?)"
	rows, err := sr.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors by status: %w", err)
	}
	defer rows.Close()

	var sensors []gen.Sensor
	for rows.Next() {
		sensor, err := scanSensorRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

// GetSensorsWithRetention returns all sensors that have a custom retention_hours set.
func (sr *SensorRepository) GetSensorsWithRetention(ctx context.Context) ([]gen.Sensor, error) {
	query := "SELECT id, name, external_id, sensor_driver, config, health_status, health_reason, enabled, status, retention_hours FROM sensors WHERE retention_hours IS NOT NULL"
	rows, err := sr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors with custom retention: %w", err)
	}
	defer rows.Close()

	var sensors []gen.Sensor
	for rows.Next() {
		sensor, err := scanSensorRow(rows)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

func (sr *SensorRepository) UpdateSensorStatus(ctx context.Context, sensorId int, status string) error {
	query := "UPDATE sensors SET status = ?, enabled = CASE WHEN ? = 'active' THEN 1 ELSE enabled END WHERE id = ?"
	result, err := sr.db.ExecContext(ctx, query, status, status, sensorId)
	if err != nil {
		return fmt.Errorf("error updating sensor status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("sensor with id %d not found", sensorId)
	}
	return nil
}
