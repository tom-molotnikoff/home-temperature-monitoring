package service

import "example/sensorHub/types"

type TemperatureServiceInterface interface {
	ServiceCollectAllSensorReadings() ([]types.APITempReading, error)
	ServiceCollectSensorReading(sensorName string) (*types.APITempReading, error)
	ServiceGetBetweenDates(table, start, end string) ([]types.APITempReading, error)
	ServiceGetLatest() ([]types.APITempReading, error)
}
