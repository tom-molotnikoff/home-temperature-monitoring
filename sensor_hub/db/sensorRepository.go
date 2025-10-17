package database

import (
	"database/sql"
	"errors"
	"example/sensorHub/types"
	"fmt"
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

func (s *SensorRepository) GetSensorIdByName(sensorName string) (int, error) {
	query := "SELECT id FROM sensors WHERE name = ?"
	var sensorID int
	err := s.db.QueryRow(query, sensorName).Scan(&sensorID)
	if err != nil {
		return 0, fmt.Errorf("could not find sensor id for name %s: %w", sensorName, err)
	}
	return sensorID, nil
}

func (s *SensorRepository) DeleteSensorByName(name string, purge bool) error {
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

	if purge {
		purgeQuery := "DELETE FROM temperature_readings WHERE sensor_id = ?"
		_, err := txn.Exec(purgeQuery, sensorId)
		if err != nil {
			return fmt.Errorf("error purging temperature readings for sensor ID %d: %w", sensorId, err)
		}
		hourlyReadingsPurgeQuery := "DELETE FROM hourly_temperature_readings WHERE sensor_id = ?"
		_, err = txn.Exec(hourlyReadingsPurgeQuery, sensorId)
		if err != nil {
			return fmt.Errorf("error purging hourly temperature readings for sensor ID %d: %w", sensorId, err)
		}
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
	query := "SELECT id, name, type, url FROM sensors WHERE type = ?"
	rows, err := s.db.Query(query, sensorType)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors by type: %w", err)
	}
	defer rows.Close()

	var sensors []types.Sensor
	for rows.Next() {
		var sensor types.Sensor
		if err := rows.Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL); err != nil {
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

	query := "INSERT INTO sensors (name, type, url) VALUES (?, ?, ?)"
	_, err := s.db.Exec(query, sensor.Name, sensor.Type, sensor.URL)
	if err != nil {
		return fmt.Errorf("error adding new sensor: %w", err)
	}
	return nil
}

func (s *SensorRepository) GetSensorByName(name string) (*types.Sensor, error) {
	query := "SELECT id, name, type, url FROM sensors WHERE name = ?"
	var sensor types.Sensor
	err := s.db.QueryRow(query, name).Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no sensor found with name %s", name)
		}
		return nil, fmt.Errorf("error querying sensor by name: %w", err)
	}
	return &sensor, nil
}

func (s *SensorRepository) GetAllSensors() ([]types.Sensor, error) {
	query := "SELECT id, name, type, url FROM sensors"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying all sensors: %w", err)
	}
	defer rows.Close()

	var sensors []types.Sensor
	for rows.Next() {
		var sensor types.Sensor
		if err := rows.Scan(&sensor.Id, &sensor.Name, &sensor.Type, &sensor.URL); err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sensor rows: %w", err)
	}
	return sensors, nil
}
