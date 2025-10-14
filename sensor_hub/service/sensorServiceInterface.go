package service

import "example/sensorHub/types"

type SensorServiceInterface interface {
	ServiceAddSensor(sensor types.Sensor) error
	ServiceUpdateSensorByName(sensor types.Sensor) error
	ServiceDeleteSensorByName(name string) error
	ServiceGetSensorByName(name string) (*types.Sensor, error)
	ServiceGetAllSensors() ([]types.Sensor, error)
	ServiceGetSensorsByType(sensorType string) ([]types.Sensor, error)
	ServiceGetSensorIdByName(name string) (int, error)
	ServiceSensorExists(name string) (bool, error)
	ServiceCollectAndStoreAllSensorReadings() error
	ServiceCollectFromSensorByName(sensorName string) error
}
