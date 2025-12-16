package database

import (
	"database/sql"
	"errors"
	"example/sensorHub/types"
	"fmt"
	"log"
)

type SensorRepository struct {
	db *sql.DB
}

func NewSensorRepository(db *sql.DB) *SensorRepository {
	return &SensorRepository{db: db}
}

func (s *SensorRepository) SensorExists(name string) (bool, error) {
	query := "SELECT COUNT(1) FROM sensors WHERE name = ?"
	var count int
	err := s.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking if sensor exists: %w", err)
	}
	return count > 0, nil
}

func (s *SensorRepository) SetEnabledSensorByName(name string, enabled bool) error {
	query := "UPDATE sensors SET enabled = ? WHERE name = ?"
	if !enabled {
		go func(name string, status types.SensorHealthStatus) {
			sensorId, err := s.GetSensorIdByName(name)
			if err != nil {
				log.Printf("failed to get sensor id for health history insert: %v", err)
				return
			}
			if sensorId <= 0 {
				log.Printf("skipping sensor health history insert: invalid sensor id %d", sensorId)
				return
			}
			insertQuery := fmt.Sprintf("INSERT INTO %s (sensor_id, health_status) VALUES (?, ?)", types.TableSensorHealthHistory)
			if _, err := s.db.Exec(insertQuery, sensorId, status); err != nil {
				log.Printf("failed to insert sensor health history for sensor %d: %v", sensorId, err)
			}
		}(name, types.SensorUnknownHealth)
		query = "UPDATE sensors SET enabled = ?, health_status = ?, health_reason = 'unknown' WHERE name = ?"
	}
	result, err := s.db.Exec(query, enabled, types.SensorUnknownHealth, name)
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

func (s *SensorRepository) GetSensorIdByName(sensorName string) (int, error) {
	query := "SELECT id FROM sensors WHERE name = ?"
	var sensorID int
	err := s.db.QueryRow(query, sensorName).Scan(&sensorID)
	if err != nil {
		return 0, fmt.Errorf("could not find sensor id for name %s: %w", sensorName, err)
	}
	return sensorID, nil
}

func (s *SensorRepository) DeleteSensorByName(name string) error {
	sensorId, err := s.GetSensorIdByName(name)
	if err != nil {
		return fmt.Errorf("error retrieving sensor ID for deletion: %w", err)
	}

	/*
	 TODO: transaction is good but purge should be its own service
	 so as to not hold up the API call whilst potentially deleting
	 a lot of data. eg: schedule a purge and let the other service
	 handle it asynchronously.
	*/

	txn, err := s.db.Begin()
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

	purgeQuery := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ?", types.TableTemperatureReadings)
	_, err = txn.Exec(purgeQuery, sensorId)
	if err != nil {
		return fmt.Errorf("error purging temperature readings for sensor ID %d: %w", sensorId, err)
	}
	hourlyReadingsPurgeQuery := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ?", types.TableHourlyAverageTemperature)
	_, err = txn.Exec(hourlyReadingsPurgeQuery, sensorId)
	if err != nil {
		return fmt.Errorf("error purging hourly temperature readings for sensor ID %d: %w", sensorId, err)
	}

	query := "DELETE FROM sensors WHERE name = ?"
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

func (s *SensorRepository) GetSensorsByType(sensorType string) ([]types.Sensor, error) {
	query := "SELECT id, name, type, url, health_status, health_reason, enabled FROM sensors WHERE type = ?"
	rows, err := s.db.Query(query, sensorType)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors by type: %w", err)
	}
	defer rows.Close()

	var sensors []types.Sensor
	for rows.Next() {
		var sensor types.Sensor
		if err := rows.Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL, &sensor.HealthStatus, &sensor.HealthReason, &sensor.Enabled); err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

func (s *SensorRepository) UpdateSensorById(sensor types.Sensor) error {
	query := "UPDATE sensors SET name = ?, type = ?, url = ? WHERE id = ?"
	result, err := s.db.Exec(query, sensor.Name, sensor.Type, sensor.URL, sensor.Id)
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

func (s *SensorRepository) AddSensor(sensor types.Sensor) error {
	if sensor.Name == "" || sensor.Type == "" || sensor.URL == "" {
		return fmt.Errorf("sensor name, type, and url cannot be empty")
	}

	query := "INSERT INTO sensors (name, type, url, health_reason, enabled) VALUES (?, ?, ?, 'unknown', ?)"
	_, err := s.db.Exec(query, sensor.Name, sensor.Type, sensor.URL, true)
	if err != nil {
		return fmt.Errorf("error adding new sensor: %w", err)
	}
	return nil
}

func (s *SensorRepository) GetSensorByName(name string) (*types.Sensor, error) {
	query := "SELECT id, name, type, url, health_status, health_reason, enabled FROM sensors WHERE name = ?"
	var sensor types.Sensor
	err := s.db.QueryRow(query, name).Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL, &sensor.HealthStatus, &sensor.HealthReason, &sensor.Enabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no sensor found with name %s", name)
		}
		return nil, fmt.Errorf("error querying sensor by name: %w", err)
	}
	return &sensor, nil
}

func (s *SensorRepository) GetAllSensors() ([]types.Sensor, error) {
	query := "SELECT id, name, type, url, health_status, health_reason, enabled FROM sensors"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying all sensors: %w", err)
	}
	defer rows.Close()

	var sensors []types.Sensor
	for rows.Next() {
		var sensor types.Sensor
		if err := rows.Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL, &sensor.HealthStatus, &sensor.HealthReason, &sensor.Enabled); err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}

func (s *SensorRepository) UpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string) error {
	query := "UPDATE sensors SET health_status = ?, health_reason = ? WHERE id = ?"
	_, err := s.db.Exec(query, healthStatus, healthReason, sensorId)
	if err != nil {
		return fmt.Errorf("error updating sensor health status: %w", err)
	}

	go func(id int, status types.SensorHealthStatus) {
		if id <= 0 {
			log.Printf("skipping sensor health history insert: invalid sensor id %d", id)
			return
		}
		insertQuery := fmt.Sprintf("INSERT INTO %s (sensor_id, health_status) VALUES (?, ?)", types.TableSensorHealthHistory)
		if _, err := s.db.Exec(insertQuery, id, status); err != nil {
			log.Printf("failed to insert sensor health history for sensor %d: %v", id, err)
		}
	}(sensorId, healthStatus)

	return nil
}

// TODO - implement methods for getting sensor health over time for reporting - see V5__sensor_health_history.sql
