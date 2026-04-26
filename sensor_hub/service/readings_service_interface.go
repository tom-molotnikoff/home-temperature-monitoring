package service

import (
	"context"
	"example/sensorHub/types"
	gen "example/sensorHub/gen"
)

type ReadingsServiceInterface interface {
	ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*types.AggregatedReadingsResponse, error)
	ServiceGetLatest(ctx context.Context) ([]gen.Reading, error)
}
