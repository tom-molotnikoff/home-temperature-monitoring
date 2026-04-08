package database

import (
	"context"
	"example/sensorHub/types"
	"time"
)

type ReadingsRepository interface {
	Add(ctx context.Context, readings []types.Reading) error
	GetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error)
	GetLatest(ctx context.Context) ([]types.Reading, error)
	GetTotalReadingsBySensorId(ctx context.Context, sensorId int) (int, error)
	DeleteReadingsOlderThan(ctx context.Context, cutoffDate time.Time) error
	ComputeHourlyAverages(ctx context.Context) error
	ComputeHourlyEvents(ctx context.Context) error
}
