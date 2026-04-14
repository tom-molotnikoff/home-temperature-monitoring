package service

import (
	"context"
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"fmt"
	"log/slog"
	"time"
)

type ReadingsService struct {
	repo    database.ReadingsRepository
	mtRepo  database.MeasurementTypeRepository
	tiers   []appProps.AggregationTier
	enabled bool
	logger  *slog.Logger
}

func NewReadingsService(repo database.ReadingsRepository, mtRepo database.MeasurementTypeRepository, tiers []appProps.AggregationTier, enabled bool, logger *slog.Logger) *ReadingsService {
	return &ReadingsService{
		repo:    repo,
		mtRepo:  mtRepo,
		tiers:   tiers,
		enabled: enabled,
		logger:  logger.With("component", "readings_service"),
	}
}

func (s *ReadingsService) ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*types.AggregatedReadingsResponse, error) {
	var interval types.AggregationInterval
	var aggFunc types.AggregationFunction

	if !s.enabled {
		interval = types.AggregationRaw
		aggFunc = types.AggregationFunctionNone
	} else if overrideInterval != "" {
		interval = types.AggregationInterval(overrideInterval)
		aggFunc = s.resolveFunction(ctx, measurementType, overrideFunction)
	} else {
		span, err := computeSpan(startDate, endDate)
		if err != nil {
			return nil, err
		}
		resolved := appProps.ResolveAggregationInterval(span, s.tiers)
		interval = types.AggregationInterval(resolved)
		aggFunc = s.resolveFunction(ctx, measurementType, overrideFunction)
	}

	if interval == types.AggregationRaw {
		aggFunc = types.AggregationFunctionNone
	}

	readings, err := s.repo.GetBetweenDates(ctx, startDate, endDate, sensorName, measurementType, interval, aggFunc)
	if err != nil {
		return nil, err
	}

	return &types.AggregatedReadingsResponse{
		AggregationInterval: interval,
		AggregationFunction: aggFunc,
		Readings:            readings,
	}, nil
}

func (s *ReadingsService) resolveFunction(ctx context.Context, measurementType, overrideFunction string) types.AggregationFunction {
	if overrideFunction != "" {
		return types.AggregationFunction(overrideFunction)
	}
	if measurementType == "" {
		return types.AggregationFunctionAvg
	}
	agg, err := s.mtRepo.GetAggregationsForMeasurementType(ctx, measurementType)
	if err != nil {
		s.logger.Warn("failed to look up aggregation function, defaulting to avg",
			"measurement_type", measurementType, "error", err)
		return types.AggregationFunctionAvg
	}
	return types.AggregationFunction(agg.DefaultFunction)
}

func (s *ReadingsService) ServiceGetLatest(ctx context.Context) ([]types.Reading, error) {
	readings, err := s.repo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}
	return readings, nil
}

func computeSpan(startDate, endDate string) (time.Duration, error) {
	const layout = "2006-01-02 15:04:05"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return 0, fmt.Errorf("parse start date: %w", err)
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		return 0, fmt.Errorf("parse end date: %w", err)
	}
	return end.Sub(start), nil
}
