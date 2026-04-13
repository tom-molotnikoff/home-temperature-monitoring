package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"example/sensorHub/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupCleanupService() (*cleanupService, *MockSensorRepository, *MockReadingsRepository, *MockFailedLoginRepository) {
	sensorRepo := new(MockSensorRepository)
	readingsRepo := new(MockReadingsRepository)
	failedRepo := new(MockFailedLoginRepository)

	service := &cleanupService{
		sensorRepo:   sensorRepo,
		readingsRepo: readingsRepo,
		failedRepo:   failedRepo,
		logger:       slog.Default().With("component", "cleanup_service"),
	}
	return service, sensorRepo, readingsRepo, failedRepo
}

// ============================================================================
// performCleanup tests
// ============================================================================

func TestCleanupService_PerformCleanup_AllEnabled(t *testing.T) {
	service, sensorRepo, readingsRepo, failedRepo := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{}).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.NoError(t, err)
	readingsRepo.AssertExpectations(t)
	sensorRepo.AssertExpectations(t)
	failedRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_AllDisabled(t *testing.T) {
	service, _, _, _ := setupCleanupService()

	// Zero values mean no cleanup should happen
	err := service.performCleanup(context.Background(), 0, 0, 0)

	assert.NoError(t, err)
}

func TestCleanupService_PerformCleanup_OnlyTemperature(t *testing.T) {
	service, sensorRepo, readingsRepo, _ := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{}).Return(nil)

	err := service.performCleanup(context.Background(), 0, 30, 0)

	assert.NoError(t, err)
	readingsRepo.AssertExpectations(t)
	sensorRepo.AssertNotCalled(t, "DeleteHealthHistoryOlderThan")
}

func TestCleanupService_PerformCleanup_OnlyHealthHistory(t *testing.T) {
	service, sensorRepo, _, _ := setupCleanupService()

	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(context.Background(), 30, 0, 0)

	assert.NoError(t, err)
	sensorRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_OnlyFailedLogins(t *testing.T) {
	service, _, _, failedRepo := setupCleanupService()

	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(context.Background(), 0, 0, 7)

	assert.NoError(t, err)
	failedRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_TemperatureError_GetSensors(t *testing.T) {
	service, sensorRepo, _, _ := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return(nil, errors.New("database error"))

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestCleanupService_PerformCleanup_TemperatureError_GlobalDelete(t *testing.T) {
	service, sensorRepo, readingsRepo, _ := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{}).Return(errors.New("database error"))

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestCleanupService_PerformCleanup_HealthHistoryError(t *testing.T) {
	service, sensorRepo, readingsRepo, _ := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{}).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(errors.New("health history error"))

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health history error")
}

func TestCleanupService_PerformCleanup_FailedLoginError(t *testing.T) {
	service, sensorRepo, readingsRepo, failedRepo := setupCleanupService()

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{}).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(errors.New("failed login error"))

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed login error")
}

func TestCleanupService_PerformCleanup_RetentionDaysCalculation(t *testing.T) {
	service, sensorRepo, readingsRepo, failedRepo := setupCleanupService()

	now := time.Now()
	expectedTempThreshold := now.AddDate(0, 0, -90)
	expectedHealthThreshold := now.AddDate(0, 0, -30)
	expectedFailedThreshold := now.AddDate(0, 0, -7)

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{}, nil)

	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return t.Sub(expectedTempThreshold).Abs() < time.Second
	}), []int{}).Return(nil)

	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return t.Sub(expectedHealthThreshold).Abs() < time.Second
	})).Return(nil)

	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		return t.Sub(expectedFailedThreshold).Abs() < time.Second
	})).Return(nil)

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.NoError(t, err)
}

func TestCleanupService_PerformCleanup_PerSensorRetention(t *testing.T) {
	service, sensorRepo, readingsRepo, failedRepo := setupCleanupService()

	retentionHours := 48
	customSensor := types.Sensor{Id: 42, Name: "sensor-co2", RetentionHours: &retentionHours}

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{customSensor}, nil)
	readingsRepo.On("DeleteReadingsOlderThanForSensor", mock.Anything, mock.MatchedBy(func(t time.Time) bool {
		expected := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
		return t.Sub(expected).Abs() < time.Second
	}), 42).Return(nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{42}).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.NoError(t, err)
	readingsRepo.AssertCalled(t, "DeleteReadingsOlderThanForSensor", mock.Anything, mock.AnythingOfType("time.Time"), 42)
	readingsRepo.AssertCalled(t, "DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{42})
}

func TestCleanupService_PerformCleanup_PerSensorRetentionError(t *testing.T) {
	service, sensorRepo, readingsRepo, _ := setupCleanupService()

	retentionHours := 24
	customSensor := types.Sensor{Id: 7, Name: "sensor-co", RetentionHours: &retentionHours}

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return([]types.Sensor{customSensor}, nil)
	readingsRepo.On("DeleteReadingsOlderThanForSensor", mock.Anything, mock.AnythingOfType("time.Time"), 7).Return(errors.New("per-sensor delete error"))

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "per-sensor delete error")
}

func TestCleanupService_PerformCleanup_MultipleCustomSensors(t *testing.T) {
	service, sensorRepo, readingsRepo, failedRepo := setupCleanupService()

	h1, h2 := 24, 720
	sensors := []types.Sensor{
		{Id: 1, Name: "sensor-a", RetentionHours: &h1},
		{Id: 2, Name: "sensor-b", RetentionHours: &h2},
	}

	sensorRepo.On("GetSensorsWithRetention", mock.Anything).Return(sensors, nil)
	readingsRepo.On("DeleteReadingsOlderThanForSensor", mock.Anything, mock.AnythingOfType("time.Time"), 1).Return(nil)
	readingsRepo.On("DeleteReadingsOlderThanForSensor", mock.Anything, mock.AnythingOfType("time.Time"), 2).Return(nil)
	readingsRepo.On("DeleteReadingsOlderThanExcludingSensors", mock.Anything, mock.AnythingOfType("time.Time"), []int{1, 2}).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(context.Background(), 30, 90, 7)

	assert.NoError(t, err)
	readingsRepo.AssertExpectations(t)
}

// ============================================================================
// NewCleanupService tests
// ============================================================================

func TestNewCleanupService_ReturnsService(t *testing.T) {
	sensorRepo := new(MockSensorRepository)
	readingsRepo := new(MockReadingsRepository)
	failedRepo := new(MockFailedLoginRepository)

	service := NewCleanupService(sensorRepo, readingsRepo, failedRepo, nil, slog.Default())

	assert.NotNil(t, service)
}
