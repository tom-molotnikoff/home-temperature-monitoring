package service

import (
	"context"
	"example/sensorHub/types"
)

type ReadingsServiceInterface interface {
	ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*types.AggregatedReadingsResponse, error)
	ServiceGetLatest(ctx context.Context) ([]types.Reading, error)
}
