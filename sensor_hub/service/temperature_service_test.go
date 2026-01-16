package service

import (
	"errors"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupTemperatureService() (*TemperatureService, *MockTemperatureRepository) {
	repo := new(MockTemperatureRepository)
	service := NewTemperatureService(repo)
	return service, repo
}

// ============================================================================
// ServiceGetBetweenDates tests
// ============================================================================

func TestTemperatureService_ServiceGetBetweenDates_Success(t *testing.T) {
	service, repo := setupTemperatureService()

	readings := []types.TemperatureReading{
		{Id: 1, SensorName: "LivingRoom", Temperature: 22.5, Time: "2025-01-15 10:00:00"},
		{Id: 2, SensorName: "LivingRoom", Temperature: 23.0, Time: "2025-01-15 11:00:00"},
		{Id: 3, SensorName: "LivingRoom", Temperature: 22.8, Time: "2025-01-15 12:00:00"},
	}
	repo.On("GetBetweenDates", "LivingRoom", "2025-01-15", "2025-01-16").Return(readings, nil)

	result, err := service.ServiceGetBetweenDates("LivingRoom", "2025-01-15", "2025-01-16")

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 22.5, result[0].Temperature)
	assert.Equal(t, "LivingRoom", result[0].SensorName)
}

func TestTemperatureService_ServiceGetBetweenDates_Empty(t *testing.T) {
	service, repo := setupTemperatureService()

	repo.On("GetBetweenDates", "Bedroom", "2025-01-15", "2025-01-16").Return([]types.TemperatureReading{}, nil)

	result, err := service.ServiceGetBetweenDates("Bedroom", "2025-01-15", "2025-01-16")

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestTemperatureService_ServiceGetBetweenDates_Error(t *testing.T) {
	service, repo := setupTemperatureService()

	repo.On("GetBetweenDates", "Unknown", "2025-01-15", "2025-01-16").Return([]types.TemperatureReading{}, errors.New("database error"))

	result, err := service.ServiceGetBetweenDates("Unknown", "2025-01-15", "2025-01-16")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)
}

func TestTemperatureService_ServiceGetBetweenDates_MultipleReadings(t *testing.T) {
	service, repo := setupTemperatureService()

	readings := make([]types.TemperatureReading, 100)
	for i := 0; i < 100; i++ {
		readings[i] = types.TemperatureReading{
			Id:          i + 1,
			SensorName:  "LivingRoom",
			Temperature: 20.0 + float64(i%10)*0.5,
			Time:        "2025-01-15 10:00:00",
		}
	}
	repo.On("GetBetweenDates", "LivingRoom", "2025-01-01", "2025-01-31").Return(readings, nil)

	result, err := service.ServiceGetBetweenDates("LivingRoom", "2025-01-01", "2025-01-31")

	assert.NoError(t, err)
	assert.Len(t, result, 100)
}

// ============================================================================
// ServiceGetLatest tests
// ============================================================================

func TestTemperatureService_ServiceGetLatest_Success(t *testing.T) {
	service, repo := setupTemperatureService()

	readings := []types.TemperatureReading{
		{Id: 10, SensorName: "LivingRoom", Temperature: 22.5, Time: "2025-01-15 14:00:00"},
		{Id: 20, SensorName: "Bedroom", Temperature: 20.0, Time: "2025-01-15 14:00:00"},
		{Id: 30, SensorName: "Kitchen", Temperature: 23.5, Time: "2025-01-15 14:00:00"},
	}
	repo.On("GetLatest").Return(readings, nil)

	result, err := service.ServiceGetLatest()

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "LivingRoom", result[0].SensorName)
	assert.Equal(t, "Bedroom", result[1].SensorName)
	assert.Equal(t, "Kitchen", result[2].SensorName)
}

func TestTemperatureService_ServiceGetLatest_Empty(t *testing.T) {
	service, repo := setupTemperatureService()

	repo.On("GetLatest").Return([]types.TemperatureReading{}, nil)

	result, err := service.ServiceGetLatest()

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestTemperatureService_ServiceGetLatest_Error(t *testing.T) {
	service, repo := setupTemperatureService()

	repo.On("GetLatest").Return([]types.TemperatureReading{}, errors.New("database error"))

	result, err := service.ServiceGetLatest()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, result)
}

// ============================================================================
// NewTemperatureService tests
// ============================================================================

func TestNewTemperatureService_ReturnsService(t *testing.T) {
	repo := new(MockTemperatureRepository)
	service := NewTemperatureService(repo)

	assert.NotNil(t, service)
}
