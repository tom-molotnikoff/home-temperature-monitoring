package service

import (
	database "example/sensorHub/db"
	"example/sensorHub/sensors"
	"example/sensorHub/types"
	"example/sensorHub/utils"
)

type TemperatureService struct {
	repo database.Repository[types.DbTempReading]
}

func NewTemperatureService(repo database.Repository[types.DbTempReading]) *TemperatureService {
	return &TemperatureService{repo: repo}
}

func (s *TemperatureService) ServiceCollectAllSensorReadings() ([]types.APITempReading, error) {
	return sensors.GetReadingFromAllTemperatureSensors()
}

func (s *TemperatureService) ServiceCollectSensorReading(sensorName string) (*types.APITempReading, error) {
	return sensors.GetReadingFromTemperatureSensor(sensorName)
}

func (s *TemperatureService) ServiceGetBetweenDates(tableName string, startDate string, endDate string) ([]types.APITempReading, error) {
	dbReadings, err := s.repo.GetBetweenDates(tableName, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return utils.ConvertDbReadingsToApiReadings(dbReadings), nil
}

func (s *TemperatureService) ServiceGetLatest() ([]types.APITempReading, error) {
	dbReadings, err := s.repo.GetLatest()
	if err != nil {
		return nil, err
	}
	return utils.ConvertDbReadingsToApiReadings(dbReadings), nil
}
