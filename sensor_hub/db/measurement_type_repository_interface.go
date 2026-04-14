package database

import (
	"context"

	"example/sensorHub/types"
)

type MeasurementTypeRepository interface {
	GetAll(ctx context.Context) ([]types.MeasurementType, error)
	GetAllWithReadings(ctx context.Context) ([]types.MeasurementType, error)
	GetByName(ctx context.Context, name string) (*types.MeasurementType, error)
	GetBySensorId(ctx context.Context, sensorId int) ([]types.SensorMeasurementType, error)
	GetMeasurementTypesWithReadings(ctx context.Context, sensorId int) ([]types.MeasurementType, error)
	EnsureExists(ctx context.Context, mt types.MeasurementType) error
	AssignToSensor(ctx context.Context, sensorId, measurementTypeId int, unit string) error
	RemoveFromSensor(ctx context.Context, sensorId, measurementTypeId int) error
	GetAggregationsForMeasurementType(ctx context.Context, name string) (*types.MeasurementTypeAggregation, error)
}
