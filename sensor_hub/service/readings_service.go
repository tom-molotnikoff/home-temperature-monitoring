package service

import (
	"context"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"log/slog"
)

type ReadingsService struct {
	repo   database.ReadingsRepository
	logger *slog.Logger
}

func NewReadingsService(repo database.ReadingsRepository, logger *slog.Logger) *ReadingsService {
	return &ReadingsService{repo: repo, logger: logger.With("component", "readings_service")}
}

func (s *ReadingsService) ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
	readings, err := s.repo.GetBetweenDates(ctx, startDate, endDate, sensorName, measurementType, hourly)
	if err != nil {
		return nil, err
	}
	return readings, nil
}

func (s *ReadingsService) ServiceGetLatest(ctx context.Context) ([]types.Reading, error) {
	readings, err := s.repo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}
	return readings, nil
}
