package database

import (
	"example/sensorHub/alerting"
	"example/sensorHub/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) UpdateLastAlertSent(ruleID int) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ruleID, sensorID, reason, numericValue, statusValue)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAllAlertRules() ([]alerting.AlertRule, error) {
	args := m.Called()
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error) {
	args := m.Called(sensorName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) CreateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockAlertRepository) UpdateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockAlertRepository) DeleteAlertRule(sensorID int) error {
	args := m.Called(sensorID)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	args := m.Called(sensorID, limit)
	return args.Get(0).([]types.AlertHistoryEntry), args.Error(1)
}

func TestMockAlertRepository(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	rule := &alerting.AlertRule{
		ID:             1,
		SensorID:       5,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		Enabled:        true,
		RateLimitHours: 1,
	}

	mockRepo.On("GetAlertRuleBySensorID", 5).Return(rule, nil)

	result, err := mockRepo.GetAlertRuleBySensorID(5)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, 5, result.SensorID)
	assert.Equal(t, alerting.AlertTypeNumericRange, result.AlertType)
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_RecordAlertSent(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	mockRepo.On("RecordAlertSent", 1, 5, "temperature too high", 35.5, "").Return(nil)

	err := mockRepo.RecordAlertSent(1, 5, "temperature too high", 35.5, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_GetAllAlertRules(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	expectedRules := []alerting.AlertRule{
		{SensorID: 1, SensorName: "Sensor1"},
		{SensorID: 2, SensorName: "Sensor2"},
	}

	mockRepo.On("GetAllAlertRules").Return(expectedRules, nil)

	rules, err := mockRepo.GetAllAlertRules()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(rules))
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_CreateAlertRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	newRule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitHours: 1,
		Enabled:        true,
	}

	mockRepo.On("CreateAlertRule", newRule).Return(nil)

	err := mockRepo.CreateAlertRule(newRule)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_UpdateAlertRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	updatedRule := &alerting.AlertRule{
		SensorID:       1,
		HighThreshold:  35.0,
		LowThreshold:   12.0,
		RateLimitHours: 2,
		Enabled:        false,
	}

	mockRepo.On("UpdateAlertRule", updatedRule).Return(nil)

	err := mockRepo.UpdateAlertRule(updatedRule)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_DeleteAlertRule(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	mockRepo.On("DeleteAlertRule", 1).Return(nil)

	err := mockRepo.DeleteAlertRule(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockAlertRepository_GetAlertHistory(t *testing.T) {
	mockRepo := new(MockAlertRepository)

	expectedHistory := []types.AlertHistoryEntry{
		{SensorID: 1, AlertType: "numeric_range", ReadingValue: "35.5"},
	}

	mockRepo.On("GetAlertHistory", 1, 10).Return(expectedHistory, nil)

	history, err := mockRepo.GetAlertHistory(1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(history))
	mockRepo.AssertExpectations(t)
}
