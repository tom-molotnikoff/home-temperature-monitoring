package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Test helpers
// ============================================================================

func setupCleanupService() (*cleanupService, *MockSensorRepository, *MockTemperatureRepository, *MockFailedLoginRepository) {
	sensorRepo := new(MockSensorRepository)
	tempRepo := new(MockTemperatureRepository)
	failedRepo := new(MockFailedLoginRepository)

	service := &cleanupService{
		sensorRepo:      sensorRepo,
		temperatureRepo: tempRepo,
		failedRepo:      failedRepo,
	}
	return service, sensorRepo, tempRepo, failedRepo
}

// ============================================================================
// performCleanup tests
// ============================================================================

func TestCleanupService_PerformCleanup_AllEnabled(t *testing.T) {
	service, sensorRepo, tempRepo, failedRepo := setupCleanupService()

	tempRepo.On("DeleteReadingsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(30, 90, 7)

	assert.NoError(t, err)
	tempRepo.AssertExpectations(t)
	sensorRepo.AssertExpectations(t)
	failedRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_AllDisabled(t *testing.T) {
	service, _, _, _ := setupCleanupService()

	// Zero values mean no cleanup should happen
	err := service.performCleanup(0, 0, 0)

	assert.NoError(t, err)
}

func TestCleanupService_PerformCleanup_OnlyTemperature(t *testing.T) {
	service, sensorRepo, tempRepo, _ := setupCleanupService()

	tempRepo.On("DeleteReadingsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	// Health history not cleaned when retention is 0
	// Failed logins not cleaned when retention is 0

	err := service.performCleanup(0, 30, 0)

	assert.NoError(t, err)
	tempRepo.AssertExpectations(t)
	sensorRepo.AssertNotCalled(t, "DeleteHealthHistoryOlderThan")
}

func TestCleanupService_PerformCleanup_OnlyHealthHistory(t *testing.T) {
	service, sensorRepo, _, _ := setupCleanupService()

	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(30, 0, 0)

	assert.NoError(t, err)
	sensorRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_OnlyFailedLogins(t *testing.T) {
	service, _, _, failedRepo := setupCleanupService()

	failedRepo.On("DeleteAttemptsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)

	err := service.performCleanup(0, 0, 7)

	assert.NoError(t, err)
	failedRepo.AssertExpectations(t)
}

func TestCleanupService_PerformCleanup_TemperatureError(t *testing.T) {
	service, _, tempRepo, _ := setupCleanupService()

	tempRepo.On("DeleteReadingsOlderThan", mock.AnythingOfType("time.Time")).Return(errors.New("database error"))

	err := service.performCleanup(30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestCleanupService_PerformCleanup_HealthHistoryError(t *testing.T) {
	service, sensorRepo, tempRepo, _ := setupCleanupService()

	tempRepo.On("DeleteReadingsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.AnythingOfType("time.Time")).Return(errors.New("health history error"))

	err := service.performCleanup(30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health history error")
}

func TestCleanupService_PerformCleanup_FailedLoginError(t *testing.T) {
	service, sensorRepo, tempRepo, failedRepo := setupCleanupService()

	tempRepo.On("DeleteReadingsOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.AnythingOfType("time.Time")).Return(nil)
	failedRepo.On("DeleteAttemptsOlderThan", mock.AnythingOfType("time.Time")).Return(errors.New("failed login error"))

	err := service.performCleanup(30, 90, 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed login error")
}

func TestCleanupService_PerformCleanup_RetentionDaysCalculation(t *testing.T) {
	service, sensorRepo, tempRepo, failedRepo := setupCleanupService()

	now := time.Now()
	expectedTempThreshold := now.AddDate(0, 0, -90)
	expectedHealthThreshold := now.AddDate(0, 0, -30)
	expectedFailedThreshold := now.AddDate(0, 0, -7)

	tempRepo.On("DeleteReadingsOlderThan", mock.MatchedBy(func(t time.Time) bool {
		// Check that threshold is within 1 second of expected
		return t.Sub(expectedTempThreshold).Abs() < time.Second
	})).Return(nil)

	sensorRepo.On("DeleteHealthHistoryOlderThan", mock.MatchedBy(func(t time.Time) bool {
		return t.Sub(expectedHealthThreshold).Abs() < time.Second
	})).Return(nil)

	failedRepo.On("DeleteAttemptsOlderThan", mock.MatchedBy(func(t time.Time) bool {
		return t.Sub(expectedFailedThreshold).Abs() < time.Second
	})).Return(nil)

	err := service.performCleanup(30, 90, 7)

	assert.NoError(t, err)
}

// ============================================================================
// NewCleanupService tests
// ============================================================================

func TestNewCleanupService_ReturnsService(t *testing.T) {
	sensorRepo := new(MockSensorRepository)
	tempRepo := new(MockTemperatureRepository)
	failedRepo := new(MockFailedLoginRepository)

	service := NewCleanupService(sensorRepo, tempRepo, failedRepo)

	assert.NotNil(t, service)
}
