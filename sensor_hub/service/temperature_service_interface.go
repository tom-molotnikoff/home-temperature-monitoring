package service

import (
	"context"
	"example/sensorHub/types"
)

type TemperatureServiceInterface interface {
	ServiceGetBetweenDates(ctx context.Context, table, start, end string) ([]types.TemperatureReading, error)
	ServiceGetLatest(ctx context.Context) ([]types.TemperatureReading, error)
}
