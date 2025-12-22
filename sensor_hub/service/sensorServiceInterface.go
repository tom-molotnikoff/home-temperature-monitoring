package service

import "example/sensorHub/types"

type SensorServiceInterface interface {
	ServiceAddSensor(sensor types.Sensor) error
	ServiceUpdateSensorById(sensor types.Sensor) error
	ServiceDeleteSensorByName(name string) error
	ServiceGetSensorByName(name string) (*types.Sensor, error)
	ServiceGetAllSensors() ([]types.Sensor, error)
	ServiceGetSensorsByType(sensorType string) ([]types.Sensor, error)
	ServiceGetSensorIdByName(name string) (int, error)
	ServiceSensorExists(name string) (bool, error)
	ServiceCollectAndStoreAllSensorReadings() error
	ServiceCollectFromSensorByName(sensorName string) error
	ServiceCollectReadingToValidateSensor(sensor types.Sensor) error
	ServiceCollectAndStoreTemperatureReadings() error
	ServiceStartPeriodicSensorCollection()
	ServiceDiscoverSensors() error
	ServiceFetchTemperatureReadingFromSensor(sensor types.Sensor) (types.TemperatureReading, error)
	ServiceFetchAllTemperatureReadings() ([]types.TemperatureReading, error)
	ServiceValidateSensorConfig(sensor types.Sensor) error
	ServiceUpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string)
	ServiceSetEnabledSensorByName(name string, enabled bool) error
	ServiceGetSensorHealthHistoryByName(name string, limit int) ([]types.SensorHealthHistory, error)
}
