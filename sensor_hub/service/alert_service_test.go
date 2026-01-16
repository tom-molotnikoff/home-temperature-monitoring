package service

import (
	"example/sensorHub/alerting"
	"example/sensorHub/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAlertRepositoryForService struct {
	mock.Mock
}

func (m *mockAlertRepositoryForService) GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *mockAlertRepositoryForService) UpdateLastAlertSent(ruleID int) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *mockAlertRepositoryForService) RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ruleID, sensorID, reason, numericValue, statusValue)
	return args.Error(0)
}

func (m *mockAlertRepositoryForService) GetAllAlertRules() ([]alerting.AlertRule, error) {
	args := m.Called()
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *mockAlertRepositoryForService) GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error) {
	args := m.Called(sensorName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *mockAlertRepositoryForService) CreateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockAlertRepositoryForService) UpdateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockAlertRepositoryForService) DeleteAlertRule(sensorID int) error {
	args := m.Called(sensorID)
	return args.Error(0)
}

func (m *mockAlertRepositoryForService) GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	args := m.Called(sensorID, limit)
	return args.Get(0).([]types.AlertHistoryEntry), args.Error(1)
}

func TestServiceGetAllAlertRules(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	expectedRules := []alerting.AlertRule{
		{SensorID: 1, SensorName: "Sensor1", AlertType: alerting.AlertTypeNumericRange},
		{SensorID: 2, SensorName: "Sensor2", AlertType: alerting.AlertTypeStatusBased},
	}

	mockRepo.On("GetAllAlertRules").Return(expectedRules, nil)

	rules, err := service.ServiceGetAllAlertRules()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(rules))
	assert.Equal(t, "Sensor1", rules[0].SensorName)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetAlertRuleBySensorID(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	expectedRule := &alerting.AlertRule{
		SensorID:   1,
		SensorName: "TestSensor",
		AlertType:  alerting.AlertTypeNumericRange,
	}

	mockRepo.On("GetAlertRuleBySensorID", 1).Return(expectedRule, nil)

	rule, err := service.ServiceGetAlertRuleBySensorID(1)

	assert.NoError(t, err)
	assert.Equal(t, "TestSensor", rule.SensorName)
	mockRepo.AssertExpectations(t)
}

func TestServiceCreateAlertRule(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	newRule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitHours: 1,
		Enabled:        true,
	}

	mockRepo.On("CreateAlertRule", newRule).Return(nil)

	err := service.ServiceCreateAlertRule(newRule)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceUpdateAlertRule(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	updatedRule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  35.0,
		LowThreshold:   12.0,
		RateLimitHours: 2,
		Enabled:        false,
	}

	mockRepo.On("UpdateAlertRule", updatedRule).Return(nil)

	err := service.ServiceUpdateAlertRule(updatedRule)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceDeleteAlertRule(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	mockRepo.On("DeleteAlertRule", 1).Return(nil)

	err := service.ServiceDeleteAlertRule(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetAlertHistory(t *testing.T) {
	mockRepo := new(mockAlertRepositoryForService)
	service := NewAlertManagementService(mockRepo)

	expectedHistory := []types.AlertHistoryEntry{
		{SensorID: 1, AlertType: "numeric_range", ReadingValue: "35.5", SentAt: time.Now()},
		{SensorID: 1, AlertType: "numeric_range", ReadingValue: "40.0", SentAt: time.Now().Add(-2 * time.Hour)},
	}

	mockRepo.On("GetAlertHistory", 1, 10).Return(expectedHistory, nil)

	history, err := service.ServiceGetAlertHistory(1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(history))
	assert.Equal(t, 1, history[0].SensorID)
	mockRepo.AssertExpectations(t)
}
