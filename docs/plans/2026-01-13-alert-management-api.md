# Alert Management API Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Create a RESTful API for managing sensor alert rules with proper RBAC permissions.

**Architecture:** Following existing API patterns (sensorApi.go, temperatureApi.go) with dedicated alertApi.go, alertRoutes.go, and AlertService. Add new RBAC permissions (view_alerts, manage_alerts) via V15 database migration. TDD approach with unit tests for all handlers.

**Tech Stack:** Go, Gin framework, existing middleware (AuthRequired, RequirePermission), testify for testing

---

## Task 1: Create RBAC Permissions for Alerts (Database Migration)

**Files:**
- Create: `sensor_hub/db/changesets/V15__add_alert_permissions.sql`

**Step 1: Create migration file**

Create the SQL migration following V12 pattern:

```sql
-- V15: Alert management permissions

INSERT IGNORE INTO permissions (name, description) VALUES ('view_alerts', 'View sensor alert rules and history');
INSERT IGNORE INTO permissions (name, description) VALUES ('manage_alerts', 'Create, update, and delete alert rules');

-- Ensure admin role has all permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin';
```

**Step 2: Verify SQL syntax**

Check that INSERT IGNORE is consistent with existing migrations (V9, V11, V12).

**Step 3: Commit**

```bash
git add sensor_hub/db/changesets/V15__add_alert_permissions.sql
git commit -m "feat(db): add RBAC permissions for alert management

- Add view_alerts permission for viewing rules and history
- Add manage_alerts permission for CRUD operations
- Grant both permissions to admin role"
```

---

## Task 2: Create Alert Service Interface (TDD)

**Files:**
- Create: `sensor_hub/service/alert_service_interface.go`
- Create: `sensor_hub/service/alert_service_test.go`

**Step 1: Write the failing test**

Create test file with mock repository:

```go
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

func (m *mockAlertRepositoryForService) RecordAlertSent(sensorID int, alertType alerting.AlertType, value string) error {
	args := m.Called(sensorID, alertType, value)
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
```

**Step 2: Run test to verify it fails**

```bash
cd sensor_hub
go test ./service/... -v -run TestService
```

Expected: FAIL - undefined: NewAlertManagementService, AlertManagementServiceInterface

**Step 3: Create interface definition**

```go
package service

import (
	"example/sensorHub/alerting"
	"example/sensorHub/types"
)

type AlertManagementServiceInterface interface {
	ServiceGetAllAlertRules() ([]alerting.AlertRule, error)
	ServiceGetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error)
	ServiceCreateAlertRule(rule *alerting.AlertRule) error
	ServiceUpdateAlertRule(rule *alerting.AlertRule) error
	ServiceDeleteAlertRule(sensorID int) error
	ServiceGetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error)
}
```

**Step 4: Create minimal service implementation**

Add to `sensor_hub/service/alert_service.go`:

```go
package service

import (
	"example/sensorHub/alerting"
	database "example/sensorHub/db"
	"example/sensorHub/types"
)

type AlertManagementService struct {
	alertRepo database.AlertRepository
}

func NewAlertManagementService(alertRepo database.AlertRepository) AlertManagementServiceInterface {
	return &AlertManagementService{
		alertRepo: alertRepo,
	}
}

func (s *AlertManagementService) ServiceGetAllAlertRules() ([]alerting.AlertRule, error) {
	return s.alertRepo.GetAllAlertRules()
}

func (s *AlertManagementService) ServiceGetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	return s.alertRepo.GetAlertRuleBySensorID(sensorID)
}

func (s *AlertManagementService) ServiceCreateAlertRule(rule *alerting.AlertRule) error {
	return s.alertRepo.CreateAlertRule(rule)
}

func (s *AlertManagementService) ServiceUpdateAlertRule(rule *alerting.AlertRule) error {
	return s.alertRepo.UpdateAlertRule(rule)
}

func (s *AlertManagementService) ServiceDeleteAlertRule(sensorID int) error {
	return s.alertRepo.DeleteAlertRule(sensorID)
}

func (s *AlertManagementService) ServiceGetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	return s.alertRepo.GetAlertHistory(sensorID, limit)
}
```

**Step 5: Run test to verify it fails (missing repo methods)**

```bash
cd sensor_hub
go test ./service/... -v -run TestService
```

Expected: FAIL - undefined methods on AlertRepository

**Step 6: Add missing methods to AlertRepository interface**

Modify `sensor_hub/db/alert_repository.go` to add interface methods and implementations. This will be done in Task 3.

---

## Task 3: Extend AlertRepository with CRUD Operations (TDD)

**Files:**
- Modify: `sensor_hub/db/alert_repository.go`
- Modify: `sensor_hub/db/alert_repository_test.go`
- Create: `sensor_hub/types/alert_history_entry.go`

**Step 1: Create AlertHistoryEntry type**

```go
package types

import "time"

type AlertHistoryEntry struct {
	ID           int       `json:"id"`
	SensorID     int       `json:"sensor_id"`
	AlertType    string    `json:"alert_type"`
	ReadingValue string    `json:"reading_value"`
	SentAt       time.Time `json:"sent_at"`
}
```

**Step 2: Write failing tests for new repository methods**

Add to `sensor_hub/db/alert_repository_test.go`:

```go
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
```

**Step 3: Run tests to verify they fail**

```bash
cd sensor_hub
go test ./db/... -v -run TestMockAlertRepository
```

Expected: FAIL - methods not defined on MockAlertRepository

**Step 4: Update AlertRepository interface and mock**

Modify `sensor_hub/db/alert_repository.go`:

```go
// Add to AlertRepository interface
type AlertRepository interface {
	GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error)
	RecordAlertSent(sensorID int, alertType alerting.AlertType, value string) error
	GetAllAlertRules() ([]alerting.AlertRule, error)
	GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error)
	CreateAlertRule(rule *alerting.AlertRule) error
	UpdateAlertRule(rule *alerting.AlertRule) error
	DeleteAlertRule(sensorID int) error
	GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error)
}

// Add implementations
func (r *alertRepository) GetAllAlertRules() ([]alerting.AlertRule, error) {
	query := `
		SELECT 
			sar.sensor_id,
			s.name,
			sar.alert_type,
			sar.high_threshold,
			sar.low_threshold,
			sar.status_trigger,
			sar.enabled,
			sar.rate_limit_hours,
			ash.sent_at
		FROM sensor_alert_rules sar
		INNER JOIN sensor s ON sar.sensor_id = s.id
		LEFT JOIN (
			SELECT sensor_id, MAX(sent_at) as sent_at
			FROM alert_sent_history
			GROUP BY sensor_id
		) ash ON sar.sensor_id = ash.sensor_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []alerting.AlertRule
	for rows.Next() {
		var rule alerting.AlertRule
		var lastAlertSentAt *time.Time
		err := rows.Scan(
			&rule.SensorID,
			&rule.SensorName,
			&rule.AlertType,
			&rule.HighThreshold,
			&rule.LowThreshold,
			&rule.TriggerStatus,
			&rule.Enabled,
			&rule.RateLimitHours,
			&lastAlertSentAt,
		)
		if err != nil {
			return nil, err
		}
		rule.LastAlertSentAt = lastAlertSentAt
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *alertRepository) GetAlertRuleBySensorName(sensorName string) (*alerting.AlertRule, error) {
	query := `
		SELECT 
			sar.sensor_id,
			s.name,
			sar.alert_type,
			sar.high_threshold,
			sar.low_threshold,
			sar.status_trigger,
			sar.enabled,
			sar.rate_limit_hours,
			ash.sent_at
		FROM sensor_alert_rules sar
		INNER JOIN sensor s ON sar.sensor_id = s.id
		LEFT JOIN (
			SELECT sensor_id, MAX(sent_at) as sent_at
			FROM alert_sent_history
			WHERE sensor_id = (SELECT id FROM sensor WHERE name = ?)
			GROUP BY sensor_id
		) ash ON sar.sensor_id = ash.sensor_id
		WHERE s.name = ?
	`

	var rule alerting.AlertRule
	var lastAlertSentAt *time.Time
	err := r.db.QueryRow(query, sensorName, sensorName).Scan(
		&rule.SensorID,
		&rule.SensorName,
		&rule.AlertType,
		&rule.HighThreshold,
		&rule.LowThreshold,
		&rule.TriggerStatus,
		&rule.Enabled,
		&rule.RateLimitHours,
		&lastAlertSentAt,
	)

	if err != nil {
		return nil, err
	}

	rule.LastAlertSentAt = lastAlertSentAt
	return &rule, nil
}

func (r *alertRepository) CreateAlertRule(rule *alerting.AlertRule) error {
	query := `
		INSERT INTO sensor_alert_rules 
		(sensor_id, alert_type, high_threshold, low_threshold, status_trigger, rate_limit_hours, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		rule.SensorID,
		rule.AlertType,
		rule.HighThreshold,
		rule.LowThreshold,
		rule.TriggerStatus,
		rule.RateLimitHours,
		rule.Enabled,
	)

	return err
}

func (r *alertRepository) UpdateAlertRule(rule *alerting.AlertRule) error {
	query := `
		UPDATE sensor_alert_rules
		SET alert_type = ?,
			high_threshold = ?,
			low_threshold = ?,
			status_trigger = ?,
			rate_limit_hours = ?,
			enabled = ?
		WHERE sensor_id = ?
	`

	_, err := r.db.Exec(query,
		rule.AlertType,
		rule.HighThreshold,
		rule.LowThreshold,
		rule.TriggerStatus,
		rule.RateLimitHours,
		rule.Enabled,
		rule.SensorID,
	)

	return err
}

func (r *alertRepository) DeleteAlertRule(sensorID int) error {
	query := `DELETE FROM sensor_alert_rules WHERE sensor_id = ?`
	_, err := r.db.Exec(query, sensorID)
	return err
}

func (r *alertRepository) GetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	query := `
		SELECT id, sensor_id, alert_type, reading_value, sent_at
		FROM alert_sent_history
		WHERE sensor_id = ?
		ORDER BY sent_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, sensorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []types.AlertHistoryEntry
	for rows.Next() {
		var entry types.AlertHistoryEntry
		err := rows.Scan(&entry.ID, &entry.SensorID, &entry.AlertType, &entry.ReadingValue, &entry.SentAt)
		if err != nil {
			return nil, err
		}
		history = append(history, entry)
	}

	return history, nil
}
```

**Step 5: Update MockAlertRepository**

Add to `sensor_hub/db/alert_repository_test.go`:

```go
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
```

**Step 6: Run tests to verify they pass**

```bash
cd sensor_hub
go test ./db/... -v -run TestMockAlertRepository
go test ./service/... -v -run TestService
```

Expected: PASS (all tests)

**Step 7: Commit**

```bash
git add sensor_hub/db/alert_repository.go sensor_hub/db/alert_repository_test.go sensor_hub/types/alert_history_entry.go sensor_hub/service/alertService.go sensor_hub/service/alertServiceInterface.go sensor_hub/service/alertService_test.go
git commit -m "feat(service): add AlertManagementService with CRUD operations

- Add AlertManagementServiceInterface with 6 methods
- Extend AlertRepository with GetAll, Create, Update, Delete, GetHistory
- Add AlertHistoryEntry type for history responses
- Add comprehensive unit tests (12 tests passing)"
```

---

## Task 4: Create Alert API Handlers (TDD)

**Files:**
- Create: `sensor_hub/api/alert_api.go`
- Create: `sensor_hub/api/alert_api_test.go`

**Step 1: Write failing tests for API handlers**

Create `sensor_hub/api/alert_api_test.go`:

```go
package api

import (
	"bytes"
	"encoding/json"
	"example/sensorHub/alerting"
	"example/sensorHub/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAlertManagementService struct {
	mock.Mock
}

func (m *mockAlertManagementService) ServiceGetAllAlertRules() ([]alerting.AlertRule, error) {
	args := m.Called()
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceGetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceCreateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceUpdateAlertRule(rule *alerting.AlertRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceDeleteAlertRule(sensorID int) error {
	args := m.Called(sensorID)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceGetAlertHistory(sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	args := m.Called(sensorID, limit)
	return args.Get(0).([]types.AlertHistoryEntry), args.Error(1)
}

func TestGetAllAlertRulesHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	expectedRules := []alerting.AlertRule{
		{SensorID: 1, SensorName: "Sensor1", AlertType: alerting.AlertTypeNumericRange, HighThreshold: 30.0, LowThreshold: 10.0},
		{SensorID: 2, SensorName: "Sensor2", AlertType: alerting.AlertTypeStatusBased, TriggerStatus: "open"},
	}

	mockService.On("ServiceGetAllAlertRules").Return(expectedRules, nil)

	router := setupTestRouter("/alerts", getAllAlertRulesHandler)
	req := httptest.NewRequest("GET", "/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Sensor1")
	assert.Contains(t, w.Body.String(), "Sensor2")
	mockService.AssertExpectations(t)
}

func TestGetAllAlertRulesHandler_Error(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	mockService.On("ServiceGetAllAlertRules").Return([]alerting.AlertRule{}, fmt.Errorf("database error"))

	router := setupTestRouter("/alerts", getAllAlertRulesHandler)
	req := httptest.NewRequest("GET", "/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Error fetching alert rules")
	mockService.AssertExpectations(t)
}

func TestGetAlertRuleBySensorIDHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	expectedRule := &alerting.AlertRule{
		SensorID:       1,
		SensorName:     "TestSensor",
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitHours: 1,
		Enabled:        true,
	}

	mockService.On("ServiceGetAlertRuleBySensorID", 1).Return(expectedRule, nil)

	router := gin.New()
	router.GET("/alerts/:sensorId", getAlertRuleBySensorIDHandler)

	req := httptest.NewRequest("GET", "/alerts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "TestSensor")
	mockService.AssertExpectations(t)
}

func TestGetAlertRuleBySensorIDHandler_InvalidID(t *testing.T) {
	router := gin.New()
	router.GET("/alerts/:sensorId", getAlertRuleBySensorIDHandler)

	req := httptest.NewRequest("GET", "/alerts/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid sensor ID")
}

func TestGetAlertRuleBySensorIDHandler_NotFound(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	mockService.On("ServiceGetAlertRuleBySensorID", 999).Return(nil, fmt.Errorf("not found"))

	router := gin.New()
	router.GET("/alerts/:sensorId", getAlertRuleBySensorIDHandler)

	req := httptest.NewRequest("GET", "/alerts/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	newRule := alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitHours: 1,
		Enabled:        true,
	}

	mockService.On("ServiceCreateAlertRule", mock.AnythingOfType("*alerting.AlertRule")).Return(nil)

	router := setupTestRouter("/alerts", createAlertRuleHandler)

	body, _ := json.Marshal(newRule)
	req := httptest.NewRequest("POST", "/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule created successfully")
	mockService.AssertExpectations(t)
}

func TestCreateAlertRuleHandler_InvalidJSON(t *testing.T) {
	router := setupTestRouter("/alerts", createAlertRuleHandler)

	req := httptest.NewRequest("POST", "/alerts", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestCreateAlertRuleHandler_ValidationError(t *testing.T) {
	router := setupTestRouter("/alerts", createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:      1,
		AlertType:     alerting.AlertTypeNumericRange,
		HighThreshold: 10.0, // Invalid: lower than low threshold
		LowThreshold:  30.0,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestUpdateAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	updatedRule := alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  35.0,
		LowThreshold:   12.0,
		RateLimitHours: 2,
		Enabled:        false,
	}

	mockService.On("ServiceUpdateAlertRule", mock.AnythingOfType("*alerting.AlertRule")).Return(nil)

	router := gin.New()
	router.PUT("/alerts/:sensorId", updateAlertRuleHandler)

	body, _ := json.Marshal(updatedRule)
	req := httptest.NewRequest("PUT", "/alerts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule updated successfully")
	mockService.AssertExpectations(t)
}

func TestDeleteAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	mockService.On("ServiceDeleteAlertRule", 1).Return(nil)

	router := gin.New()
	router.DELETE("/alerts/:sensorId", deleteAlertRuleHandler)

	req := httptest.NewRequest("DELETE", "/alerts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule deleted successfully")
	mockService.AssertExpectations(t)
}

func TestGetAlertHistoryHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	expectedHistory := []types.AlertHistoryEntry{
		{SensorID: 1, AlertType: "numeric_range", ReadingValue: "35.5", SentAt: time.Now()},
		{SensorID: 1, AlertType: "numeric_range", ReadingValue: "40.0", SentAt: time.Now().Add(-2 * time.Hour)},
	}

	mockService.On("ServiceGetAlertHistory", 1, 10).Return(expectedHistory, nil)

	router := gin.New()
	router.GET("/alerts/:sensorId/history", getAlertHistoryHandler)

	req := httptest.NewRequest("GET", "/alerts/1/history?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "35.5")
	assert.Contains(t, w.Body.String(), "numeric_range")
	mockService.AssertExpectations(t)
}

func TestGetAlertHistoryHandler_DefaultLimit(t *testing.T) {
	mockService := new(mockAlertManagementService)
	alertManagementService = mockService

	mockService.On("ServiceGetAlertHistory", 1, 50).Return([]types.AlertHistoryEntry{}, nil)

	router := gin.New()
	router.GET("/alerts/:sensorId/history", getAlertHistoryHandler)

	req := httptest.NewRequest("GET", "/alerts/1/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd sensor_hub
go test ./api/... -v -run TestAlert
```

Expected: FAIL - undefined handlers and alertManagementService

**Step 3: Create API handlers**

Create `sensor_hub/api/alert_api.go`:

```go
package api

import (
	"example/sensorHub/alerting"
	"example/sensorHub/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var alertManagementService service.AlertManagementServiceInterface

func InitAlertAPI(s service.AlertManagementServiceInterface) {
	alertManagementService = s
}

func getAllAlertRulesHandler(ctx *gin.Context) {
	rules, err := alertManagementService.ServiceGetAllAlertRules()
	if err != nil {
		log.Printf("Error fetching alert rules: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert rules", "error": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, rules)
}

func getAlertRuleBySensorIDHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	rule, err := alertManagementService.ServiceGetAlertRuleBySensorID(sensorID)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "Alert rule not found", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, rule)
}

func createAlertRuleHandler(ctx *gin.Context) {
	var rule alerting.AlertRule
	if err := ctx.BindJSON(&rule); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Validate the rule
	if err := rule.Validate(); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceCreateAlertRule(&rule); err != nil {
		log.Printf("Error creating alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, gin.H{"message": "Alert rule created successfully"})
}

func updateAlertRuleHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	var rule alerting.AlertRule
	if err := ctx.BindJSON(&rule); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Ensure the sensor ID from URL matches the rule
	rule.SensorID = sensorID

	// Validate the rule
	if err := rule.Validate(); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid alert rule", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceUpdateAlertRule(&rule); err != nil {
		log.Printf("Error updating alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule updated successfully"})
}

func deleteAlertRuleHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	if err := alertManagementService.ServiceDeleteAlertRule(sensorID); err != nil {
		log.Printf("Error deleting alert rule: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting alert rule", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Alert rule deleted successfully"})
}

func getAlertHistoryHandler(ctx *gin.Context) {
	sensorIDStr := ctx.Param("sensorId")
	sensorID, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid sensor ID", "error": err.Error()})
		return
	}

	// Default limit is 50, max 100
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	history, err := alertManagementService.ServiceGetAlertHistory(sensorID, limit)
	if err != nil {
		log.Printf("Error fetching alert history: %v", err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching alert history", "error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, history)
}
```

**Step 4: Run tests to verify they pass**

```bash
cd sensor_hub
go test ./api/... -v -run TestAlert
```

Expected: PASS (all 13 tests)

**Step 5: Commit**

```bash
git add sensor_hub/api/alertApi.go sensor_hub/api/alert_api_test.go
git commit -m "feat(api): add alert management API handlers with tests

- Add getAllAlertRulesHandler (GET /alerts)
- Add getAlertRuleBySensorIDHandler (GET /alerts/:sensorId)
- Add createAlertRuleHandler (POST /alerts)
- Add updateAlertRuleHandler (PUT /alerts/:sensorId)
- Add deleteAlertRuleHandler (DELETE /alerts/:sensorId)
- Add getAlertHistoryHandler (GET /alerts/:sensorId/history)
- Add 13 comprehensive unit tests"
```

---

## Task 5: Create Alert Routes with RBAC

**Files:**
- Create: `sensor_hub/api/alert_routes.go`

**Step 1: Create route registration following existing pattern**

```go
package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAlertRoutes(router *gin.Engine) {
	alertsGroup := router.Group("/alerts")
	{
		// View alert rules and history
		alertsGroup.GET("/", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAllAlertRulesHandler)
		alertsGroup.GET("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertRuleBySensorIDHandler)
		alertsGroup.GET("/:sensorId/history", middleware.AuthRequired(), middleware.RequirePermission("view_alerts"), getAlertHistoryHandler)

		// Manage alert rules (create, update, delete)
		alertsGroup.POST("/", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), createAlertRuleHandler)
		alertsGroup.PUT("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), updateAlertRuleHandler)
		alertsGroup.DELETE("/:sensorId", middleware.AuthRequired(), middleware.RequirePermission("manage_alerts"), deleteAlertRuleHandler)
	}
}
```

**Step 2: Verify route structure matches existing patterns**

Check against sensorRoutes.go and temperatureRoutes.go for consistency.

**Step 3: Commit**

```bash
git add sensor_hub/api/alert_routes.go
git commit -m "feat(api): add alert management routes with RBAC

- Register /alerts endpoints with permission middleware
- GET routes require view_alerts permission
- POST/PUT/DELETE routes require manage_alerts permission
- Follows existing API route patterns"
```

---

## Task 6: Wire Alert API into Main Application

**Files:**
- Modify: `sensor_hub/main.go`

**Step 1: Add alert service initialization**

Find the service initialization section (around line 46) and add:

```go
alertManagementService := service.NewAlertManagementService(alertRepo)
```

**Step 2: Add API initialization**

Find the API initialization section (around line 56) and add:

```go
api.InitAlertAPI(alertManagementService)
```

**Step 3: Register routes**

Find the route registration section (after InitAuthMiddleware) and add:

```go
api.RegisterAlertRoutes(router)
```

**Step 4: Build application to verify wiring**

```bash
cd sensor_hub
go build
```

Expected: Build succeeds with no errors

**Step 5: Commit**

```bash
git add sensor_hub/main.go
git commit -m "feat(main): wire alert management API into application

- Initialize AlertManagementService with alertRepo
- Initialize alert API with service
- Register alert routes
- Complete integration of alert management system"
```

---

## Task 7: End-to-End Verification

**Step 1: Run all tests**

```bash
cd sensor_hub
go test ./... -v 2>&1 | grep -E "(PASS|FAIL|ok|coverage)"
```

Expected: All new tests passing, integration tests may fail (require running server)

**Step 2: Check test coverage for new packages**

```bash
cd sensor_hub
go test ./api/... -cover -run TestAlert
go test ./service/... -cover -run TestService.*Alert
```

Expected: >85% coverage for new code

**Step 3: Build application**

```bash
cd sensor_hub
go build -o /tmp/sensor_hub_alert_api_test
file /tmp/sensor_hub_alert_api_test
rm /tmp/sensor_hub_alert_api_test
```

Expected: Successful build, ELF executable

**Step 4: Review all changes**

```bash
git log --oneline -7
git diff origin/main --stat
```

**Step 5: Create summary documentation**

Update `docs/alerting-system.md` with new API section:

```markdown
## API Endpoints

### Alert Rule Management

#### GET /alerts
List all alert rules for all sensors.

**Permissions:** `view_alerts`

**Response:**
```json
[
  {
    "ID": 0,
    "SensorID": 1,
    "SensorName": "Bedroom",
    "AlertType": "numeric_range",
    "HighThreshold": 25.0,
    "LowThreshold": 12.0,
    "TriggerStatus": "",
    "Enabled": true,
    "RateLimitHours": 1,
    "LastAlertSentAt": "2026-01-13T10:00:00Z"
  }
]
```

#### GET /alerts/:sensorId
Get alert rule for a specific sensor.

**Permissions:** `view_alerts`

**Response:** Single AlertRule object (404 if not found)

#### POST /alerts
Create a new alert rule.

**Permissions:** `manage_alerts`

**Request Body:**
```json
{
  "SensorID": 1,
  "AlertType": "numeric_range",
  "HighThreshold": 30.0,
  "LowThreshold": 10.0,
  "RateLimitHours": 1,
  "Enabled": true
}
```

**Validation:**
- `numeric_range`: HighThreshold must be > LowThreshold
- `status_based`: TriggerStatus must not be empty

#### PUT /alerts/:sensorId
Update an existing alert rule.

**Permissions:** `manage_alerts`

**Request Body:** Same as POST (SensorID in URL overrides body)

#### DELETE /alerts/:sensorId
Delete an alert rule for a sensor.

**Permissions:** `manage_alerts`

**Response:** 200 OK with confirmation message

#### GET /alerts/:sensorId/history
Get alert history for a sensor.

**Permissions:** `view_alerts`

**Query Parameters:**
- `limit` (optional): Number of entries to return (1-100, default 50)

**Response:**
```json
[
  {
    "id": 1,
    "sensor_id": 1,
    "alert_type": "numeric_range",
    "reading_value": "35.5",
    "sent_at": "2026-01-13T10:00:00Z"
  }
]
```

### RBAC Permissions

- **view_alerts**: View alert rules and history (read-only)
- **manage_alerts**: Create, update, and delete alert rules

Both permissions are automatically granted to the `admin` role by the V15 migration.
```

**Step 6: Commit documentation**

```bash
git add docs/alerting-system.md
git commit -m "docs: add API endpoints documentation for alert management

- Document all 6 alert management endpoints
- Add request/response examples
- Describe RBAC permissions
- Include validation rules"
```

---

## Summary

**Total Tasks:** 7
**Total Files:** 11 created, 2 modified
**Total Tests:** ~30 new tests
**Expected Coverage:** >85% for new code

**Key Deliverables:**
1. V15 database migration for RBAC permissions
2. AlertManagementService with 6 operations
3. Extended AlertRepository with CRUD methods
4. Alert API with 6 RESTful endpoints
5. Complete test coverage for service and API layers
6. Proper RBAC integration with existing middleware
7. Updated documentation

**Next Steps After This Plan:**
- Apply V15 migration using Flyway
- Test API with real HTTP requests (Postman/curl)
- Consider UI implementation (future work)
- Add integration tests if desired
