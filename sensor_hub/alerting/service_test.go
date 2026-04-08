package alerting

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) GetAlertRuleBySensorID(ctx context.Context, sensorID int) (*AlertRule, error) {
	args := m.Called(ctx, sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AlertRule), args.Error(1)
}

func (m *MockAlertRepository) UpdateLastAlertSent(ctx context.Context, ruleID int) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) RecordAlertSent(ctx context.Context, ruleID, sensorID, measurementTypeId int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ctx, ruleID, sensorID, measurementTypeId, reason, numericValue, statusValue)
	return args.Error(0)
}

func TestAlertService_ProcessReadingAlert_NoRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	mockRepo.On("GetAlertRuleBySensorID", mock.Anything, 1).Return(nil, nil)

	service := NewAlertService(mockRepo, slog.Default())
	err := service.ProcessReadingAlert(context.Background(), 1, "TestSensor", "temperature", 25.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_NumericInRange(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	rule := &AlertRule{
		ID:             1,
		SensorID:       1,
		SensorName:     "TestSensor",
		AlertType:      AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		Enabled:        true,
		RateLimitHours: 1,
	}

	mockRepo.On("GetAlertRuleBySensorID", mock.Anything, 1).Return(rule, nil)

	service := NewAlertService(mockRepo, slog.Default())
	err := service.ProcessReadingAlert(context.Background(), 1, "TestSensor", "temperature", 20.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_NumericExceedsHigh(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	callbackChan := make(chan bool, 1)

	rule := &AlertRule{
		ID:             1,
		SensorID:       1,
		SensorName:     "TestSensor",
		AlertType:      AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		Enabled:        true,
		RateLimitHours: 1,
	}

	mockRepo.On("GetAlertRuleBySensorID", mock.Anything, 1).Return(rule, nil)
	mockRepo.On("RecordAlertSent", mock.Anything, 1, 1, 0, mock.AnythingOfType("string"), 35.0, "").Return(nil)

	service := NewAlertService(mockRepo, slog.Default())
	service.SetInAppNotificationCallback(func(sensorName, sensorType, reason string, numericValue float64) {
		callbackChan <- true
	})
	err := service.ProcessReadingAlert(context.Background(), 1, "TestSensor", "temperature", 35.0, "")

	assert.NoError(t, err)
	select {
	case <-callbackChan:
		// callback was called
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not called")
	}
	mockRepo.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_RateLimited(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)

	rule := &AlertRule{
		ID:              1,
		SensorID:        1,
		SensorName:      "TestSensor",
		AlertType:       AlertTypeNumericRange,
		HighThreshold:   30.0,
		LowThreshold:    10.0,
		Enabled:         true,
		RateLimitHours:  1,
		LastAlertSentAt: &thirtyMinutesAgo,
	}

	mockRepo.On("GetAlertRuleBySensorID", mock.Anything, 1).Return(rule, nil)

	service := NewAlertService(mockRepo, slog.Default())
	err := service.ProcessReadingAlert(context.Background(), 1, "TestSensor", "temperature", 35.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_StatusBased(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	callbackChan := make(chan bool, 1)

	rule := &AlertRule{
		ID:             2,
		SensorID:       2,
		SensorName:     "DoorSensor",
		AlertType:      AlertTypeStatusBased,
		TriggerStatus:  "open",
		Enabled:        true,
		RateLimitHours: 0,
	}

	mockRepo.On("GetAlertRuleBySensorID", mock.Anything, 2).Return(rule, nil)
	mockRepo.On("RecordAlertSent", mock.Anything, 2, 2, 0, mock.AnythingOfType("string"), 0.0, "open").Return(nil)

	service := NewAlertService(mockRepo, slog.Default())
	service.SetInAppNotificationCallback(func(sensorName, sensorType, reason string, numericValue float64) {
		callbackChan <- true
	})
	err := service.ProcessReadingAlert(context.Background(), 2, "DoorSensor", "door", 0, "open")

	assert.NoError(t, err)
	select {
	case <-callbackChan:
		// callback was called
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not called")
	}
	mockRepo.AssertExpectations(t)
}
