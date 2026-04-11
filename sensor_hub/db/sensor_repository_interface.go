package database

import (
	"context"
	"time"

	"example/sensorHub/types"
)

type SensorRepositoryInterface[T any] interface {
	AddSensor(ctx context.Context, sensor T) error
	UpdateSensorById(ctx context.Context, sensor T) error
	DeleteSensorByName(ctx context.Context, name string) error
	GetSensorByName(ctx context.Context, name string) (*T, error)
	SetEnabledSensorByName(ctx context.Context, name string, enabled bool) error
	GetAllSensors(ctx context.Context) ([]T, error)
	GetSensorsByDriver(ctx context.Context, sensorDriver string) ([]T, error)
	GetSensorsByStatus(ctx context.Context, status string) ([]T, error)
	GetSensorIdByName(ctx context.Context, name string) (int, error)
	SensorExists(ctx context.Context, name string) (bool, error)
	UpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus types.SensorHealthStatus, healthReason string) error
	UpdateSensorStatus(ctx context.Context, sensorId int, status string) error
	GetSensorHealthHistoryById(ctx context.Context, sensorId int, limit int) ([]types.SensorHealthHistory, error)
	DeleteHealthHistoryOlderThan(ctx context.Context, cutoffDate time.Time) error
}
