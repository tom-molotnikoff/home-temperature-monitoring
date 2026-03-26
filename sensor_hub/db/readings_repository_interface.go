package database

import (
	"context"
	"time"
)

type ReadingsRepository[T any] interface {
	Add(ctx context.Context, readings []T) error
	GetBetweenDates(ctx context.Context, tableName, startDate, endDate, sensorName string) ([]T, error)
	GetLatest(ctx context.Context) ([]T, error)
	GetTotalReadingsBySensorId(ctx context.Context, sensorId int) (int, error)
	DeleteReadingsOlderThan(ctx context.Context, cutoffDate time.Time) error
	ComputeHourlyAverages(ctx context.Context) error
}
