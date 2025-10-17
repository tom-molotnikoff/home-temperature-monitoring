package database

import "example/sensorHub/types"

type SensorRepositoryInterface[T any] interface {
	AddSensor(sensor T) error
	UpdateSensorById(sensor T) error
	DeleteSensorByName(name string, purge bool) error
	GetSensorByName(name string) (*T, error)
	GetAllSensors() ([]T, error)
	GetSensorsByType(sensorType string) ([]T, error)
	GetSensorIdByName(name string) (int, error)
	SensorExists(name string) (bool, error)
	UpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string) error
}
