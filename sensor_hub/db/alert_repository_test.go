package database

import (
	"database/sql"
	"errors"
	"example/sensorHub/alerting"
	"example/sensorHub/types"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAlertRepository is kept for use by other packages
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

// ============================================================================
// GetAlertRuleBySensorID tests (implementation)
// ============================================================================

func TestAlertRepository_GetAlertRuleBySensorID_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	now := time.Now()
	dbMock.ExpectQuery("SELECT").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(alertRuleColumns).
			AddRow(1, 5, "sensor-1", "numeric_range", 30.0, 10.0, nil, true, 1, now))

	rule, err := repo.GetAlertRuleBySensorID(5)

	assert.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, 1, rule.ID)
	assert.Equal(t, 5, rule.SensorID)
	assert.Equal(t, alerting.AlertTypeNumericRange, rule.AlertType)
	assert.Equal(t, 30.0, rule.HighThreshold)
	assert.Equal(t, 10.0, rule.LowThreshold)
	assert.True(t, rule.Enabled)
	require.NotNil(t, rule.LastAlertSentAt)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorID_NotFound(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	rule, err := repo.GetAlertRuleBySensorID(999)

	assert.NoError(t, err)
	assert.Nil(t, rule)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorID_NullLastAlertSent(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(alertRuleColumns).
			AddRow(1, 5, "sensor-1", "numeric_range", 30.0, 10.0, nil, true, 1, nil))

	rule, err := repo.GetAlertRuleBySensorID(5)

	assert.NoError(t, err)
	require.NotNil(t, rule)
	assert.Nil(t, rule.LastAlertSentAt)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorID_WithTriggerStatus(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(alertRuleColumns).
			AddRow(1, 5, "sensor-1", "status_based", 0.0, 0.0, "bad", true, 1, nil))

	rule, err := repo.GetAlertRuleBySensorID(5)

	assert.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, alerting.AlertTypeStatusBased, rule.AlertType)
	assert.Equal(t, "bad", rule.TriggerStatus)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorID_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(5).
		WillReturnError(errors.New("database error"))

	rule, err := repo.GetAlertRuleBySensorID(5)

	assert.Error(t, err)
	assert.Nil(t, rule)
	assert.Contains(t, err.Error(), "failed to get alert rule")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// RecordAlertSent tests
// ============================================================================

func TestAlertRepository_RecordAlertSent_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("INSERT INTO alert_sent_history").
		WithArgs(1, 5, "temperature too high", 35.5, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.RecordAlertSent(1, 5, "temperature too high", 35.5, "")

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_RecordAlertSent_WithStatusValue(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("INSERT INTO alert_sent_history").
		WithArgs(1, 5, "sensor status changed", 0.0, "bad").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.RecordAlertSent(1, 5, "sensor status changed", 0.0, "bad")

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_RecordAlertSent_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("INSERT INTO alert_sent_history").
		WithArgs(1, 5, "test", 0.0, "").
		WillReturnError(errors.New("database error"))

	err := repo.RecordAlertSent(1, 5, "test", 0.0, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record alert sent")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetAllAlertRules tests
// ============================================================================

func TestAlertRepository_GetAllAlertRules_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	now := time.Now()
	dbMock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows(alertRuleColumnsNoID).
			AddRow(1, "sensor-1", "numeric_range", 30.0, 10.0, nil, true, 1, now).
			AddRow(2, "sensor-2", "status_based", 0.0, 0.0, "bad", true, 2, nil))

	rules, err := repo.GetAllAlertRules()

	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "sensor-1", rules[0].SensorName)
	assert.Equal(t, "sensor-2", rules[1].SensorName)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAllAlertRules_Empty(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows(alertRuleColumnsNoID))

	rules, err := repo.GetAllAlertRules()

	assert.NoError(t, err)
	assert.Empty(t, rules)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAllAlertRules_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WillReturnError(errors.New("database error"))

	rules, err := repo.GetAllAlertRules()

	assert.Error(t, err)
	assert.Nil(t, rules)
	assert.Contains(t, err.Error(), "failed to get all alert rules")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetAlertRuleBySensorName tests
// ============================================================================

func TestAlertRepository_GetAlertRuleBySensorName_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs("sensor-1", "sensor-1").
		WillReturnRows(sqlmock.NewRows(alertRuleColumnsNoID).
			AddRow(1, "sensor-1", "numeric_range", 30.0, 10.0, nil, true, 1, nil))

	rule, err := repo.GetAlertRuleBySensorName("sensor-1")

	assert.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "sensor-1", rule.SensorName)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorName_NotFound(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs("nonexistent", "nonexistent").
		WillReturnError(sql.ErrNoRows)

	rule, err := repo.GetAlertRuleBySensorName("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, rule)
	assert.Contains(t, err.Error(), "alert rule not found")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertRuleBySensorName_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs("sensor-1", "sensor-1").
		WillReturnError(errors.New("database error"))

	rule, err := repo.GetAlertRuleBySensorName("sensor-1")

	assert.Error(t, err)
	assert.Nil(t, rule)
	assert.Contains(t, err.Error(), "failed to get alert rule")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// CreateAlertRule tests
// ============================================================================

func TestAlertRepository_CreateAlertRule_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	rule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitHours: 1,
		Enabled:        true,
	}

	dbMock.ExpectExec("INSERT INTO sensor_alert_rules").
		WithArgs(1, alerting.AlertTypeNumericRange, 30.0, 10.0, "", 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateAlertRule(rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_CreateAlertRule_StatusBased(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	rule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeStatusBased,
		TriggerStatus:  "bad",
		RateLimitHours: 2,
		Enabled:        true,
	}

	dbMock.ExpectExec("INSERT INTO sensor_alert_rules").
		WithArgs(1, alerting.AlertTypeStatusBased, 0.0, 0.0, "bad", 2, true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateAlertRule(rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_CreateAlertRule_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	rule := &alerting.AlertRule{
		SensorID:  1,
		AlertType: alerting.AlertTypeNumericRange,
	}

	dbMock.ExpectExec("INSERT INTO sensor_alert_rules").
		WithArgs(1, alerting.AlertTypeNumericRange, 0.0, 0.0, "", 0, false).
		WillReturnError(errors.New("duplicate entry"))

	err := repo.CreateAlertRule(rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create alert rule")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// UpdateAlertRule tests
// ============================================================================

func TestAlertRepository_UpdateAlertRule_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	rule := &alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  35.0,
		LowThreshold:   12.0,
		RateLimitHours: 2,
		Enabled:        false,
	}

	dbMock.ExpectExec("UPDATE sensor_alert_rules").
		WithArgs(alerting.AlertTypeNumericRange, 35.0, 12.0, "", 2, false, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateAlertRule(rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_UpdateAlertRule_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	rule := &alerting.AlertRule{
		SensorID: 1,
	}

	dbMock.ExpectExec("UPDATE sensor_alert_rules").
		WithArgs(alerting.AlertType(""), 0.0, 0.0, "", 0, false, 1).
		WillReturnError(errors.New("database error"))

	err := repo.UpdateAlertRule(rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update alert rule")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// DeleteAlertRule tests
// ============================================================================

func TestAlertRepository_DeleteAlertRule_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("DELETE FROM sensor_alert_rules WHERE sensor_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteAlertRule(1)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_DeleteAlertRule_NotFound(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("DELETE FROM sensor_alert_rules WHERE sensor_id = \\?").
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteAlertRule(999)

	assert.NoError(t, err) // Not an error if nothing to delete
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_DeleteAlertRule_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectExec("DELETE FROM sensor_alert_rules WHERE sensor_id = \\?").
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	err := repo.DeleteAlertRule(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete alert rule")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetAlertHistory tests
// ============================================================================

func TestAlertRepository_GetAlertHistory_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	now := time.Now()
	dbMock.ExpectQuery("SELECT").
		WithArgs(1, 10).
		WillReturnRows(sqlmock.NewRows(alertHistoryColumns).
			AddRow(1, 1, "numeric_range", 35.5, now).
			AddRow(2, 1, "numeric_range", 36.0, now.Add(-time.Hour)))

	history, err := repo.GetAlertHistory(1, 10)

	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, "35.50", history[0].ReadingValue)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertHistory_Empty(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(1, 10).
		WillReturnRows(sqlmock.NewRows(alertHistoryColumns))

	history, err := repo.GetAlertHistory(1, 10)

	assert.NoError(t, err)
	assert.Empty(t, history)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertHistory_NullReadingValue(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	now := time.Now()
	dbMock.ExpectQuery("SELECT").
		WithArgs(1, 10).
		WillReturnRows(sqlmock.NewRows(alertHistoryColumns).
			AddRow(1, 1, "status_based", nil, now))

	history, err := repo.GetAlertHistory(1, 10)

	assert.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Empty(t, history[0].ReadingValue)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestAlertRepository_GetAlertHistory_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	repo := NewAlertRepository(db)

	dbMock.ExpectQuery("SELECT").
		WithArgs(1, 10).
		WillReturnError(errors.New("database error"))

	history, err := repo.GetAlertHistory(1, 10)

	assert.Error(t, err)
	assert.Nil(t, history)
	assert.Contains(t, err.Error(), "failed to get alert history")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// UpdateLastAlertSent tests (compatibility method)
// ============================================================================

func TestAlertRepository_UpdateLastAlertSent_NoOp(t *testing.T) {
	db, _ := newMockDB(t)
	repo := NewAlertRepository(db)

	// This is a no-op method kept for backwards compatibility
	err := repo.UpdateLastAlertSent(1)

	assert.NoError(t, err)
}
