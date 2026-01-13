package alerting

import (
	"errors"
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

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendAlert(sensorName, sensorType, reason string, numericValue float64, statusValue string) error {
	args := m.Called(sensorName, sensorType, reason, numericValue, statusValue)
	return args.Error(0)
}

func TestAlertService_ProcessReadingAlert_NoRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(nil, nil)

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 25.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertNotCalled(t, "SendAlert")
}

func TestAlertService_ProcessReadingAlert_NumericInRange(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

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

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 20.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertNotCalled(t, "SendAlert")
}

func TestAlertService_ProcessReadingAlert_NumericExceedsHigh(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

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
	mockNotifier.On("SendAlert", "TestSensor", "temperature", mock.AnythingOfType("string"), 35.0, "").Return(nil)
	mockRepo.On("RecordAlertSent", 1, 1, mock.AnythingOfType("string"), 35.0, "").Return(nil)

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 35.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_RateLimited(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

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

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 35.0, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertNotCalled(t, "SendAlert")
}

func TestAlertService_ProcessReadingAlert_StatusBased(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

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
	mockNotifier.On("SendAlert", "DoorSensor", "door", mock.AnythingOfType("string"), 0.0, "open").Return(nil)
	mockRepo.On("RecordAlertSent", 2, 2, mock.AnythingOfType("string"), 0.0, "open").Return(nil)

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(2, "DoorSensor", "door", 0, "open")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestAlertService_ProcessReadingAlert_NotifierFails(t *testing.T) {
	mockRepo := new(MockAlertRepository)
	mockNotifier := new(MockNotifier)

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
	mockNotifier.On("SendAlert", "TestSensor", "temperature", mock.AnythingOfType("string"), 35.0, "").
		Return(errors.New("SMTP connection failed"))

	service := NewAlertService(mockRepo, mockNotifier)
	err := service.ProcessReadingAlert(1, "TestSensor", "temperature", 35.0, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send alert")
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "RecordAlertSent")
}
