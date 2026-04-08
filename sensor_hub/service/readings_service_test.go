package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupReadingsService() (*ReadingsService, *MockReadingsRepository) {
	repo := new(MockReadingsRepository)
	service := NewReadingsService(repo, slog.Default())
	return service, repo
}

// ============================================================================
// ServiceGetBetweenDates tests
// ============================================================================

func TestReadingsService_ServiceGetBetweenDates_Success(t *testing.T) {
	service, repo := setupReadingsService()

	val1 := 22.5
	val2 := 23.0
	val3 := 22.8
	readings := []types.Reading{
		{Id: 1, SensorName: "LivingRoom", MeasurementType: "temperature", Unit: "°C", NumericValue: &val1, Time: "2025-01-15 10:00:00"},
		{Id: 2, SensorName: "LivingRoom", MeasurementType: "temperature", Unit: "°C", NumericValue: &val2, Time: "2025-01-15 11:00:00"},
		{Id: 3, SensorName: "LivingRoom", MeasurementType: "temperature", Unit: "°C", NumericValue: &val3, Time: "2025-01-15 12:00:00"},
	}
	repo.On("GetBetweenDates", mock.Anything, "2025-01-15", "2025-01-16", "", "", false).Return(readings, nil)

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-15", "2025-01-16", "", "", false)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 22.5, *result[0].NumericValue)
	assert.Equal(t, "LivingRoom", result[0].SensorName)
}

func TestReadingsService_ServiceGetBetweenDates_Empty(t *testing.T) {
	service, repo := setupReadingsService()

	repo.On("GetBetweenDates", mock.Anything, "2025-01-15", "2025-01-16", "", "", false).Return([]types.Reading{}, nil)

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-15", "2025-01-16", "", "", false)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestReadingsService_ServiceGetBetweenDates_Error(t *testing.T) {
	service, repo := setupReadingsService()

	repo.On("GetBetweenDates", mock.Anything, "2025-01-15", "2025-01-16", "", "", false).Return([]types.Reading{}, errors.New("database error"))

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-15", "2025-01-16", "", "", false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)
}

func TestReadingsService_ServiceGetBetweenDates_MultipleReadings(t *testing.T) {
	service, repo := setupReadingsService()

	readings := make([]types.Reading, 100)
	for i := 0; i < 100; i++ {
		val := 20.0 + float64(i%10)*0.5
		readings[i] = types.Reading{
			Id:              i + 1,
			SensorName:      "LivingRoom",
			MeasurementType: "temperature",
			Unit:            "°C",
			NumericValue:    &val,
			Time:            "2025-01-15 10:00:00",
		}
	}
	repo.On("GetBetweenDates", mock.Anything, "2025-01-01", "2025-01-31", "", "", false).Return(readings, nil)

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-01", "2025-01-31", "", "", false)

	assert.NoError(t, err)
	assert.Len(t, result, 100)
}

func TestReadingsService_ServiceGetBetweenDates_WithSensorFilter(t *testing.T) {
	service, repo := setupReadingsService()

	val := 22.5
	readings := []types.Reading{
		{Id: 1, SensorName: "Office", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2025-01-15 10:00:00"},
	}
	repo.On("GetBetweenDates", mock.Anything, "2025-01-15", "2025-01-16", "Office", "", false).Return(readings, nil)

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-15", "2025-01-16", "Office", "", false)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Office", result[0].SensorName)
}

func TestReadingsService_ServiceGetBetweenDates_Hourly(t *testing.T) {
	service, repo := setupReadingsService()

	val := 22.5
	readings := []types.Reading{
		{Id: 1, SensorName: "Office", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2025-01-15 10:00:00"},
	}
	repo.On("GetBetweenDates", mock.Anything, "2025-01-15", "2025-01-16", "", "", true).Return(readings, nil)

	result, err := service.ServiceGetBetweenDates(context.Background(), "2025-01-15", "2025-01-16", "", "", true)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

// ============================================================================
// ServiceGetLatest tests
// ============================================================================

func TestReadingsService_ServiceGetLatest_Success(t *testing.T) {
	service, repo := setupReadingsService()

	val1 := 22.5
	val2 := 20.0
	val3 := 23.5
	readings := []types.Reading{
		{Id: 10, SensorName: "LivingRoom", MeasurementType: "temperature", Unit: "°C", NumericValue: &val1, Time: "2025-01-15 14:00:00"},
		{Id: 20, SensorName: "Bedroom", MeasurementType: "temperature", Unit: "°C", NumericValue: &val2, Time: "2025-01-15 14:00:00"},
		{Id: 30, SensorName: "Kitchen", MeasurementType: "temperature", Unit: "°C", NumericValue: &val3, Time: "2025-01-15 14:00:00"},
	}
	repo.On("GetLatest", mock.Anything).Return(readings, nil)

	result, err := service.ServiceGetLatest(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "LivingRoom", result[0].SensorName)
	assert.Equal(t, "Bedroom", result[1].SensorName)
	assert.Equal(t, "Kitchen", result[2].SensorName)
}

func TestReadingsService_ServiceGetLatest_Empty(t *testing.T) {
	service, repo := setupReadingsService()

	repo.On("GetLatest", mock.Anything).Return([]types.Reading{}, nil)

	result, err := service.ServiceGetLatest(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestReadingsService_ServiceGetLatest_Error(t *testing.T) {
	service, repo := setupReadingsService()

	repo.On("GetLatest", mock.Anything).Return([]types.Reading{}, errors.New("database error"))

	result, err := service.ServiceGetLatest(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)
}

// ============================================================================
// NewReadingsService tests
// ============================================================================

func TestNewReadingsService_ReturnsService(t *testing.T) {
	repo := new(MockReadingsRepository)
	service := NewReadingsService(repo, slog.Default())

	assert.NotNil(t, service)
}
