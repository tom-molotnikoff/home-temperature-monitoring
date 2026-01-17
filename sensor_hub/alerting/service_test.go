package alerting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) GetAlertRuleBySensorID(sensorID int) (*AlertRule, error) {
	args := m.Called(sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AlertRule), args.Error(1)
}

func (m *MockAlertRepository) UpdateLastAlertSent(ruleID int) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ruleID, sensorID, reason, numericValue, statusValue)
	return args.Error(0)
}

func TestAlertService_ProcessReadingAlert_NoRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(nil, nil)

	service := NewAlertService(mockRepo)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 25.0, "")

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

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(rule, nil)

	service := NewAlertService(mockRepo)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 20.0, "")

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

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(rule, nil)
	mockRepo.On("RecordAlertSent", 1, 1, mock.AnythingOfType("string"), 35.0, "").Return(nil)

	service := NewAlertService(mockRepo)
	service.SetInAppNotificationCallback(func(sensorName, sensorType, reason string, numericValue float64) {
		callbackChan <- true
	})
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 35.0, "")

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

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(rule, nil)

	service := NewAlertService(mockRepo)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 35.0, "")

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

	mockRepo.On("GetAlertRuleBySensorID", 2).Return(rule, nil)
	mockRepo.On("RecordAlertSent", 2, 2, mock.AnythingOfType("string"), 0.0, "open").Return(nil)

	service := NewAlertService(mockRepo)
	service.SetInAppNotificationCallback(func(sensorName, sensorType, reason string, numericValue float64) {
		callbackChan <- true
	})
	err := service.ProcessReadingAlert(2, "DoorSensor", "door", 0, "open")

	assert.NoError(t, err)
	select {
	case <-callbackChan:
		// callback was called
	case <-time.After(100 * time.Millisecond):
		t.Fatal("callback was not called")
	}
	mockRepo.AssertExpectations(t)
}
