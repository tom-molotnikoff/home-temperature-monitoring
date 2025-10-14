package service

import "example/sensorHub/types"

type TemperatureServiceInterface interface {
	ServiceGetBetweenDates(table, start, end string) ([]types.TemperatureReading, error)
	ServiceGetLatest() ([]types.TemperatureReading, error)
}
