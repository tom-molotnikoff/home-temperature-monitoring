package service

import (
	database "example/sensorHub/db"
	"example/sensorHub/types"
)

type TemperatureService struct {
	repo database.ReadingsRepository[types.TemperatureReading]
}

func NewTemperatureService(repo database.ReadingsRepository[types.TemperatureReading]) *TemperatureService {
	return &TemperatureService{repo: repo}
}

func (s *TemperatureService) ServiceGetBetweenDates(tableName string, startDate string, endDate string) ([]types.TemperatureReading, error) {
	readings, err := s.repo.GetBetweenDates(tableName, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return readings, nil
}

func (s *TemperatureService) ServiceGetLatest() ([]types.TemperatureReading, error) {
	readings, err := s.repo.GetLatest()
	if err != nil {
		return nil, err
	}
	return readings, nil
}
