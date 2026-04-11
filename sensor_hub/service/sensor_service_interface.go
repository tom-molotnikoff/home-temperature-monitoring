package service

import (
	"context"
	"example/sensorHub/types"
)

type SensorServiceInterface interface {
	ServiceAddSensor(ctx context.Context, sensor types.Sensor) error
	ServiceUpdateSensorById(ctx context.Context, sensor types.Sensor) error
	ServiceDeleteSensorByName(ctx context.Context, name string) error
	ServiceGetSensorByName(ctx context.Context, name string) (*types.Sensor, error)
	ServiceGetSensorByExternalId(ctx context.Context, externalId string) (*types.Sensor, error)
	ServiceGetAllSensors(ctx context.Context) ([]types.Sensor, error)
	ServiceGetSensorsByDriver(ctx context.Context, sensorDriver string) ([]types.Sensor, error)
	ServiceGetSensorIdByName(ctx context.Context, name string) (int, error)
	ServiceSensorExists(ctx context.Context, name string) (bool, error)
	ServiceSensorExistsByExternalId(ctx context.Context, externalId string) (bool, error)
	ServiceCollectAndStoreAllSensorReadings(ctx context.Context) error
	ServiceCollectFromSensorByName(ctx context.Context, sensorName string) error
	ServiceCollectReadingToValidateSensor(ctx context.Context, sensor types.Sensor) error
	ServiceStartPeriodicSensorCollection(ctx context.Context)
	ServiceDiscoverSensors(ctx context.Context) error
	ServiceValidateSensorConfig(ctx context.Context, sensor types.Sensor) error
	ServiceUpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus types.SensorHealthStatus, healthReason string)
	ServiceSetEnabledSensorByName(ctx context.Context, name string, enabled bool) error
	ServiceGetSensorHealthHistoryByName(ctx context.Context, name string, limit int) ([]types.SensorHealthHistory, error)
	ServiceGetTotalReadingsForEachSensor(ctx context.Context) (map[string]int, error)
	ServiceGetSensorsByStatus(ctx context.Context, status string) ([]types.Sensor, error)
	ServiceApproveSensor(ctx context.Context, sensorId int) error
	ServiceDismissSensor(ctx context.Context, sensorId int) error
	ServiceProcessPushReadings(ctx context.Context, sensor types.Sensor, readings []types.Reading) error
	ServiceGetMeasurementTypesForSensor(ctx context.Context, sensorId int) ([]types.MeasurementType, error)
	ServiceGetAllMeasurementTypes(ctx context.Context) ([]types.MeasurementType, error)
}
