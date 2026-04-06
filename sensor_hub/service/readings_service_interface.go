package service

import (
	"context"
	"example/sensorHub/types"
)

type ReadingsServiceInterface interface {
	ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error)
	ServiceGetLatest(ctx context.Context) ([]types.Reading, error)
}
