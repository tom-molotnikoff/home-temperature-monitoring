package service

import (
	"context"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log/slog"
)

type TemperatureService struct {
	repo   database.ReadingsRepository[types.TemperatureReading]
	logger *slog.Logger
}

func NewTemperatureService(repo database.ReadingsRepository[types.TemperatureReading], logger *slog.Logger) *TemperatureService {
	return &TemperatureService{repo: repo, logger: logger.With("component", "temperature_service")}
}

func (s *TemperatureService) ServiceGetBetweenDates(ctx context.Context, tableName string, startDate string, endDate string) ([]types.TemperatureReading, error) {
	readings, err := s.repo.GetBetweenDates(ctx, tableName, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return readings, nil
}

func (s *TemperatureService) ServiceGetLatest(ctx context.Context) ([]types.TemperatureReading, error) {
	readings, err := s.repo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}
	return readings, nil
}
