package service

import "example/sensorHub/types"

type TemperatureServiceInterface interface {
	ServiceCollectAllSensorReadings() ([]types.TemperatureReading, error)
	ServiceCollectSensorReading(sensorName string) (*types.TemperatureReading, error)
	ServiceGetBetweenDates(table, start, end string) ([]types.TemperatureReading, error)
	ServiceGetLatest() ([]types.TemperatureReading, error)
}
