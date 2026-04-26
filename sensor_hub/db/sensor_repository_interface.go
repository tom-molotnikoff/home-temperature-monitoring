package database

import (
	"context"
	"time"

	gen "example/sensorHub/gen"
)

type SensorRepositoryInterface[T any] interface {
	AddSensor(ctx context.Context, sensor T) error
	UpdateSensorById(ctx context.Context, sensor T, retentionHoursPresent bool) error
	DeleteSensorByName(ctx context.Context, name string) error
	GetSensorByName(ctx context.Context, name string) (*T, error)
	GetSensorByExternalId(ctx context.Context, externalId string) (*T, error)
	GetSensorById(ctx context.Context, id int) (*T, error)
	SetEnabledSensorByName(ctx context.Context, name string, enabled bool) error
	GetAllSensors(ctx context.Context) ([]T, error)
	GetSensorsByDriver(ctx context.Context, sensorDriver string) ([]T, error)
	GetSensorsByStatus(ctx context.Context, status string) ([]T, error)
	GetSensorIdByName(ctx context.Context, name string) (int, error)
	SensorExists(ctx context.Context, name string) (bool, error)
	SensorExistsByExternalId(ctx context.Context, externalId string) (bool, error)
	UpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus gen.SensorHealthStatus, healthReason string) error
	UpdateSensorStatus(ctx context.Context, sensorId int, status string) error
	GetSensorHealthHistoryById(ctx context.Context, sensorId int, limit int) ([]gen.SensorHealthHistory, error)
	DeleteHealthHistoryOlderThan(ctx context.Context, cutoffDate time.Time) error
	GetSensorsWithRetention(ctx context.Context) ([]gen.Sensor, error)
}
