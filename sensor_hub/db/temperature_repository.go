package database

import (
	"context"
	"database/sql"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

type TemperatureRepository struct {
	db         *sql.DB
	sensorRepo SensorRepositoryInterface[types.Sensor]
	logger     *slog.Logger
}

var validTemperatureTables = map[string]struct{}{
	types.TableTemperatureReadings:      {},
	types.TableHourlyAverageTemperature: {},
}

var temperatureColumnByTable = map[string]string{
	types.TableTemperatureReadings:      "temperature",
	types.TableHourlyAverageTemperature: "average_temperature",
}

func NewTemperatureRepository(db *sql.DB, sensorRepo SensorRepositoryInterface[types.Sensor], logger *slog.Logger) *TemperatureRepository {
	return &TemperatureRepository{db: db, sensorRepo: sensorRepo, logger: logger.With("component", "temperature_repository")}
}

func (r *TemperatureRepository) Add(ctx context.Context, readings []types.TemperatureReading) error {
	query := fmt.Sprintf("INSERT INTO %s (sensor_id, time, temperature) VALUES (?, ?, ?)", types.TableTemperatureReadings)
	for _, reading := range readings {
		sensorID, err := r.sensorRepo.GetSensorIdByName(ctx, reading.SensorName)
		if err != nil {
			return fmt.Errorf("issue finding sensor id: %w", err)
		}
		_, err = r.db.ExecContext(ctx, query, sensorID, reading.Time, strconv.FormatFloat(reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %w", err)
		}
		r.logger.Debug("saved reading to database", "sensor", reading.SensorName)
	}
	return nil
}

func (r *TemperatureRepository) GetBetweenDates(ctx context.Context, tableName string, startDate string, endDate string, sensorName string) ([]types.TemperatureReading, error) {
	if _, ok := validTemperatureTables[tableName]; !ok {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	column, ok := temperatureColumnByTable[tableName]
	if !ok {
		return nil, fmt.Errorf("unknown temperature column for table: %s", tableName)
	}

	query := fmt.Sprintf(`
		SELECT tr.id, s.name AS sensor_name, tr.time, tr.%s
		FROM %s tr
		JOIN sensors s ON tr.sensor_id = s.id
		WHERE tr.time BETWEEN ? AND ?
	`, column, tableName)

	args := []any{startDate, endDate}

	if sensorName != "" {
		query += " AND LOWER(s.name) = LOWER(?)"
		args = append(args, sensorName)
	}

	query += " ORDER BY tr.time ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer func() { _ = rows.Close() }()

	readings, err := scanDbTempReading(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning readings: %w", err)
	}
	return readings, nil
}

func (r *TemperatureRepository) GetTotalReadingsBySensorId(ctx context.Context, sensorId int) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s 
		WHERE sensor_id = ?
	`, types.TableTemperatureReadings)

	var count int
	err := r.db.QueryRowContext(ctx, query, sensorId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error fetching total readings for sensor ID %d: %w", sensorId, err)
	}
	return count, nil
}

func (r *TemperatureRepository) GetLatest(ctx context.Context) ([]types.TemperatureReading, error) {
	query := fmt.Sprintf(`
		SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature
		FROM %s tr
		JOIN sensors s ON tr.sensor_id = s.id
		ORDER BY tr.time DESC
		LIMIT 30
	`, types.TableTemperatureReadings)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	readings, err := scanDbTempReading(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning readings: %w", err)
	}

	latestReadingsPerSensor := make(map[string]types.TemperatureReading)
	for _, r := range readings {
		sensorName := r.SensorName
		if _, exists := latestReadingsPerSensor[sensorName]; !exists {
			latestReadingsPerSensor[sensorName] = r
		}
	}

	finalReadings := make([]types.TemperatureReading, 0, len(latestReadingsPerSensor))
	for _, r := range latestReadingsPerSensor {
		finalReadings = append(finalReadings, r)
	}
	return finalReadings, nil
}

func (r *TemperatureRepository) DeleteReadingsOlderThan(ctx context.Context, cutoffDateTime time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf("DELETE FROM %s WHERE time < ?", types.TableTemperatureReadings)
	if _, err := tx.ExecContext(ctx, query, cutoffDateTime); err != nil {
		return fmt.Errorf("error deleting old temperature readings from %s: %w", types.TableTemperatureReadings, err)
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE time < ?", types.TableHourlyAverageTemperature)
	if _, err := tx.ExecContext(ctx, query, cutoffDateTime); err != nil {
		return fmt.Errorf("error deleting old temperature readings from %s: %w", types.TableHourlyAverageTemperature, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
func scanDbTempReading(rows *sql.Rows) ([]types.TemperatureReading, error) {
	var readings []types.TemperatureReading
	for rows.Next() {
		var reading types.TemperatureReading
		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.Time, &reading.Temperature)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		reading.Time = utils.NormalizeTimeToSpaceFormat(reading.Time)
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return readings, nil
}

func (r *TemperatureRepository) ComputeHourlyAverages(ctx context.Context) error {
	query := `
		INSERT OR IGNORE INTO hourly_avg_temperature (sensor_id, time, average_temperature)
		SELECT
			tr.sensor_id,
			strftime('%Y-%m-%d %H:00:00', tr.time) AS hour,
			ROUND(AVG(tr.temperature), 2) AS avg_temp
		FROM temperature_readings tr
		WHERE tr.time >= strftime('%Y-%m-%d %H:00:00', datetime('now', '-1 hour'))
		  AND tr.time < strftime('%Y-%m-%d %H:00:00', datetime('now'))
		GROUP BY tr.sensor_id, hour
	`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error computing hourly averages: %w", err)
	}
	return nil
}
