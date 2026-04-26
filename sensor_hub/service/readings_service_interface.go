package service

import (
	"context"
	gen "example/sensorHub/gen"
)

type ReadingsServiceInterface interface {
	ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error)
	ServiceGetLatest(ctx context.Context) ([]gen.Reading, error)
}
