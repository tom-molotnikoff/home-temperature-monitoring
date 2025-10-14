package database

import (
	"database/sql"
	"example/sensorHub/types"
	"fmt"
	"log"
	"strconv"
)

type TemperatureRepository struct {
	db *sql.DB
}

const (
	TableTemperatureReadings      = "temperature_readings"
	TableHourlyAverageTemperature = "hourly_avg_temperature"
)

var validTemperatureTables = map[string]struct{}{
	TableTemperatureReadings:      {},
	TableHourlyAverageTemperature: {},
}

func NewTemperatureRepository(db *sql.DB) *TemperatureRepository {
	return &TemperatureRepository{db: db}
}

func (r *TemperatureRepository) Add(readings []types.DbTempReading) error {
	query := fmt.Sprintf("INSERT INTO %s (sensor_name, time, temperature) VALUES (?, ?, ?)", TableTemperatureReadings)
	for _, reading := range readings {
		_, err := r.db.Exec(query, reading.SensorName, reading.Time, strconv.FormatFloat(reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %w", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.SensorName)
	}
	return nil
}

func (r *TemperatureRepository) GetBetweenDates(tableName string, startDate string, endDate string) ([]types.DbTempReading, error) {
	if _, ok := validTemperatureTables[tableName]; !ok {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE time BETWEEN ? AND ? ORDER BY time ASC", tableName)

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer rows.Close()

	readings, err := scanDbReading(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning readings: %w", err)
	}
	return readings, nil
}

func (r *TemperatureRepository) GetLatest() ([]types.DbTempReading, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY time DESC LIMIT 30", TableTemperatureReadings)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer rows.Close()

	readings, err := scanDbReading(rows)
	if err != nil {
		return nil, fmt.Errorf("error scanning readings: %w", err)
	}

	latestReadingsPerSensor := make(map[string]types.DbTempReading)
	for _, r := range readings {
		sensorName := r.SensorName
		if _, exists := latestReadingsPerSensor[sensorName]; !exists {
			latestReadingsPerSensor[sensorName] = r
		}
	}

	finalReadings := make([]types.DbTempReading, 0, len(latestReadingsPerSensor))
	for _, r := range latestReadingsPerSensor {
		finalReadings = append(finalReadings, r)
	}
	return finalReadings, nil
}
