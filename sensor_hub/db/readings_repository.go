package database

import (
	"context"
	"database/sql"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type ReadingsRepositoryImpl struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewReadingsRepository(db *sql.DB, logger *slog.Logger) ReadingsRepository {
	return &ReadingsRepositoryImpl{db: db, logger: logger.With("component", "readings_repository")}
}

func (r *ReadingsRepositoryImpl) Add(ctx context.Context, readings []types.Reading) error {
	var stored int
	for _, reading := range readings {
		mtID, err := r.resolveMeasurementTypeID(ctx, reading.MeasurementType)
		if err != nil {
			r.logger.Warn("skipping reading with unknown measurement type",
				"sensor", reading.SensorName, "type", reading.MeasurementType)
			continue
		}

		sensorID, err := r.resolveSensorID(ctx, reading.SensorName)
		if err != nil {
			return fmt.Errorf("issue finding sensor id: %w", err)
		}

		query := fmt.Sprintf("INSERT INTO %s (sensor_id, measurement_type_id, numeric_value, text_state, time) VALUES (?, ?, ?, ?, ?)", types.TableReadings)
		_, err = r.db.ExecContext(ctx, query, sensorID, mtID, reading.NumericValue, reading.TextState, reading.Time)
		if err != nil {
			return fmt.Errorf("issue persisting reading to database: %w", err)
		}
		stored++
		r.logger.Debug("saved reading to database", "sensor", reading.SensorName, "type", reading.MeasurementType)
	}
	if stored == 0 && len(readings) > 0 {
		return fmt.Errorf("no readings stored: all %d readings had unrecognised measurement types", len(readings))
	}
	return nil
}

func (r *ReadingsRepositoryImpl) GetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
	var query string
	if hourly {
		query = fmt.Sprintf(`
			SELECT ha.id, s.name, mt.name, ha.average_value, NULL, COALESCE(smt.unit, mt.default_unit), ha.time
			FROM %s ha
			JOIN sensors s ON ha.sensor_id = s.id
			JOIN %s mt ON ha.measurement_type_id = mt.id
			LEFT JOIN %s smt ON smt.sensor_id = s.id AND smt.measurement_type_id = mt.id
			WHERE ha.time BETWEEN ? AND ?
		`, types.TableHourlyAverages, types.TableMeasurementTypes, types.TableSensorMeasurementTypes)
	} else {
		query = fmt.Sprintf(`
			SELECT r.id, s.name, mt.name, r.numeric_value, r.text_state, COALESCE(smt.unit, mt.default_unit), r.time
			FROM %s r
			JOIN sensors s ON r.sensor_id = s.id
			JOIN %s mt ON r.measurement_type_id = mt.id
			LEFT JOIN %s smt ON smt.sensor_id = s.id AND smt.measurement_type_id = mt.id
			WHERE r.time BETWEEN ? AND ?
		`, types.TableReadings, types.TableMeasurementTypes, types.TableSensorMeasurementTypes)
	}

	args := []any{startDate, endDate}

	if sensorName != "" {
		query += " AND LOWER(s.name) = LOWER(?)"
		args = append(args, sensorName)
	}

	if measurementType != "" {
		query += " AND LOWER(mt.name) = LOWER(?)"
		args = append(args, measurementType)
	}

	if hourly {
		query += " ORDER BY ha.time ASC"
	} else {
		query += " ORDER BY r.time ASC"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer func() { _ = rows.Close() }()

	return scanReadings(rows)
}

func (r *ReadingsRepositoryImpl) GetLatest(ctx context.Context) ([]types.Reading, error) {
	query := fmt.Sprintf(`
		SELECT sub.id, sub.sensor_name, sub.measurement_type, sub.numeric_value, sub.text_state, sub.unit, sub.time
		FROM (
			SELECT r.id, s.name AS sensor_name, mt.name AS measurement_type,
				r.numeric_value, r.text_state, COALESCE(smt.unit, mt.default_unit) AS unit, r.time,
				ROW_NUMBER() OVER (PARTITION BY r.sensor_id, r.measurement_type_id ORDER BY r.time DESC) AS rn
			FROM %s r
			JOIN sensors s ON r.sensor_id = s.id
			JOIN %s mt ON r.measurement_type_id = mt.id
			LEFT JOIN %s smt ON smt.sensor_id = s.id AND smt.measurement_type_id = mt.id
		) sub
		WHERE sub.rn = 1
	`, types.TableReadings, types.TableMeasurementTypes, types.TableSensorMeasurementTypes)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanReadings(rows)
}

func (r *ReadingsRepositoryImpl) GetTotalReadingsBySensorId(ctx context.Context, sensorId int) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE sensor_id = ?", types.TableReadings)
	var count int
	err := r.db.QueryRowContext(ctx, query, sensorId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error fetching total readings for sensor ID %d: %w", sensorId, err)
	}
	return count, nil
}

func (r *ReadingsRepositoryImpl) DeleteReadingsOlderThan(ctx context.Context, cutoffDateTime time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, table := range []string{types.TableReadings, types.TableHourlyAverages, types.TableHourlyEvents} {
		query := fmt.Sprintf("DELETE FROM %s WHERE time < ?", table)
		if _, err := tx.ExecContext(ctx, query, cutoffDateTime); err != nil {
			return fmt.Errorf("error deleting old readings from %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// DeleteReadingsOlderThanForSensor deletes readings older than cutoff for a specific sensor.
// All three reading tables (readings, hourly_averages, hourly_events) are cleaned atomically.
func (r *ReadingsRepositoryImpl) DeleteReadingsOlderThanForSensor(ctx context.Context, cutoffDateTime time.Time, sensorId int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, table := range []string{types.TableReadings, types.TableHourlyAverages, types.TableHourlyEvents} {
		query := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ? AND time < ?", table)
		if _, err := tx.ExecContext(ctx, query, sensorId, cutoffDateTime); err != nil {
			return fmt.Errorf("error deleting old readings from %s for sensor %d: %w", table, sensorId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// DeleteReadingsOlderThanExcludingSensors deletes readings older than cutoff for all sensors
// except those in the excludedSensorIds list (which have custom per-sensor retention applied separately).
// If excludedSensorIds is empty, all sensors are cleaned (same as DeleteReadingsOlderThan).
func (r *ReadingsRepositoryImpl) DeleteReadingsOlderThanExcludingSensors(ctx context.Context, cutoffDateTime time.Time, excludedSensorIds []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, table := range []string{types.TableReadings, types.TableHourlyAverages, types.TableHourlyEvents} {
		var query string
		var args []any
		if len(excludedSensorIds) == 0 {
			query = fmt.Sprintf("DELETE FROM %s WHERE time < ?", table)
			args = []any{cutoffDateTime}
		} else {
			placeholders := strings.Repeat("?,", len(excludedSensorIds))
			placeholders = placeholders[:len(placeholders)-1] // trim trailing comma
			query = fmt.Sprintf("DELETE FROM %s WHERE time < ? AND sensor_id NOT IN (%s)", table, placeholders)
			args = make([]any, 0, 1+len(excludedSensorIds))
			args = append(args, cutoffDateTime)
			for _, id := range excludedSensorIds {
				args = append(args, id)
			}
		}
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("error deleting old readings from %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *ReadingsRepositoryImpl) ComputeHourlyAverages(ctx context.Context) error {
	query := fmt.Sprintf(`
		INSERT OR IGNORE INTO %s (sensor_id, measurement_type_id, time, average_value)
		SELECT
			r.sensor_id,
			r.measurement_type_id,
			strftime('%%Y-%%m-%%d %%H:00:00', r.time) AS hour,
			ROUND(AVG(r.numeric_value), 2)
		FROM %s r
		JOIN %s mt ON r.measurement_type_id = mt.id
		WHERE mt.category = 'numeric'
		  AND r.numeric_value IS NOT NULL
		  AND r.time >= strftime('%%Y-%%m-%%d %%H:00:00', datetime('now', '-1 hour'))
		  AND r.time < strftime('%%Y-%%m-%%d %%H:00:00', datetime('now'))
		GROUP BY r.sensor_id, r.measurement_type_id, hour
	`, types.TableHourlyAverages, types.TableReadings, types.TableMeasurementTypes)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error computing hourly averages: %w", err)
	}
	return nil
}

func (r *ReadingsRepositoryImpl) ComputeHourlyEvents(ctx context.Context) error {
	query := fmt.Sprintf(`
		INSERT OR IGNORE INTO %s (sensor_id, measurement_type_id, time, event_count)
		SELECT
			r.sensor_id,
			r.measurement_type_id,
			strftime('%%Y-%%m-%%d %%H:00:00', r.time) AS hour,
			COUNT(*)
		FROM %s r
		JOIN %s mt ON r.measurement_type_id = mt.id
		WHERE mt.category = 'binary'
		  AND r.text_state IS NOT NULL
		  AND r.time >= strftime('%%Y-%%m-%%d %%H:00:00', datetime('now', '-1 hour'))
		  AND r.time < strftime('%%Y-%%m-%%d %%H:00:00', datetime('now'))
		GROUP BY r.sensor_id, r.measurement_type_id, hour
	`, types.TableHourlyEvents, types.TableReadings, types.TableMeasurementTypes)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error computing hourly events: %w", err)
	}
	return nil
}

func (r *ReadingsRepositoryImpl) resolveMeasurementTypeID(ctx context.Context, name string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT id FROM %s WHERE LOWER(name) = LOWER(?)", types.TableMeasurementTypes), name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("measurement type %q not found: %w", name, err)
	}
	return id, nil
}

func (r *ReadingsRepositoryImpl) resolveSensorID(ctx context.Context, name string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, "SELECT id FROM sensors WHERE LOWER(name) = LOWER(?)", name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("sensor %q not found: %w", name, err)
	}
	return id, nil
}

func scanReadings(rows *sql.Rows) ([]types.Reading, error) {
	var readings []types.Reading
	for rows.Next() {
		var reading types.Reading
		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.MeasurementType, &reading.NumericValue, &reading.TextState, &reading.Unit, &reading.Time)
		if err != nil {
			return nil, fmt.Errorf("error scanning reading row: %w", err)
		}
		reading.Time = utils.NormalizeTimeToSpaceFormat(reading.Time)
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over reading rows: %w", err)
	}
	return readings, nil
}
