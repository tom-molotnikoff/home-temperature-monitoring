package database

import (
	"database/sql"
	"example/sensorHub/types"
	"fmt"
	"log"
	"strconv"
)

type TemperatureRepository struct {
	db         *sql.DB
	sensorRepo SensorRepositoryInterface[types.Sensor]
}

var validTemperatureTables = map[string]struct{}{
	types.TableTemperatureReadings:      {},
	types.TableHourlyAverageTemperature: {},
}

var temperatureColumnByTable = map[string]string{
	types.TableTemperatureReadings:      "temperature",
	types.TableHourlyAverageTemperature: "average_temperature",
}

func NewTemperatureRepository(db *sql.DB, sensorRepo SensorRepositoryInterface[types.Sensor]) *TemperatureRepository {
	return &TemperatureRepository{db: db, sensorRepo: sensorRepo}
}

func (r *TemperatureRepository) Add(readings []types.TemperatureReading) error {
	query := fmt.Sprintf("INSERT INTO %s (sensor_id, time, temperature) VALUES (?, ?, ?)", types.TableTemperatureReadings)
	for _, reading := range readings {
		sensorID, err := r.sensorRepo.GetSensorIdByName(reading.SensorName)
		if err != nil {
			return fmt.Errorf("issue finding sensor id: %w", err)
		}
		_, err = r.db.Exec(query, sensorID, reading.Time, strconv.FormatFloat(reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %w", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.SensorName)
	}
	return nil
}

func (r *TemperatureRepository) GetBetweenDates(tableName string, startDate string, endDate string) ([]types.TemperatureReading, error) {
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
		ORDER BY tr.time ASC
	`, column, tableName)

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer rows.Close()

	readings, err := scanDbTempReading(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning readings: %w", err)
	}
	return readings, nil
}

func (r *TemperatureRepository) GetLatest() ([]types.TemperatureReading, error) {
	query := fmt.Sprintf(`
		SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature
		FROM %s tr
		JOIN sensors s ON tr.sensor_id = s.id
		ORDER BY tr.time DESC
		LIMIT 30
	`, types.TableTemperatureReadings)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer rows.Close()

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

func scanDbTempReading(rows *sql.Rows) ([]types.TemperatureReading, error) {
	var readings []types.TemperatureReading
	for rows.Next() {
		var reading types.TemperatureReading
		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.Time, &reading.Temperature)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return readings, nil
}
