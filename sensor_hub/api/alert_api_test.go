package api

import (
	"bytes"
	"context"
	"encoding/json"
	"example/sensorHub/alerting"
	gen "example/sensorHub/gen"
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

func (m *mockAlertManagementService) ServiceGetAllAlertRules(ctx context.Context) ([]alerting.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceGetAlertRuleByID(ctx context.Context, ruleID int) (*alerting.AlertRule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceGetAlertRuleBySensorID(ctx context.Context, sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceGetAlertRulesBySensorID(ctx context.Context, sensorID int) ([]alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID)
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *mockAlertManagementService) ServiceCreateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceUpdateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceDeleteAlertRule(ctx context.Context, ruleID int) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *mockAlertManagementService) ServiceGetAlertHistory(ctx context.Context, sensorID int, limit int) ([]gen.AlertHistoryEntry, error) {
	args := m.Called(ctx, sensorID, limit)
	return args.Get(0).([]gen.AlertHistoryEntry), args.Error(1)
}

func TestGetAllAlertRulesHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	expectedRules := []alerting.AlertRule{
		{SensorID: 1, SensorName: "Sensor1", AlertType: alerting.AlertTypeNumericRange, HighThreshold: 30.0, LowThreshold: 10.0},
		{SensorID: 2, SensorName: "Sensor2", AlertType: alerting.AlertTypeStatusBased, TriggerStatus: "open"},
	}

	mockService.On("ServiceGetAllAlertRules", mock.Anything).Return(expectedRules, nil)

	router := setupTestRouter("/alerts", s.getAllAlertRulesHandler)
	req := httptest.NewRequest("GET", "/api/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Sensor1")
	assert.Contains(t, w.Body.String(), "Sensor2")
	mockService.AssertExpectations(t)
}

func TestGetAllAlertRulesHandler_Error(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	mockService.On("ServiceGetAllAlertRules", mock.Anything).Return([]alerting.AlertRule{}, fmt.Errorf("database error"))

	router := setupTestRouter("/alerts", s.getAllAlertRulesHandler)
	req := httptest.NewRequest("GET", "/api/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Error fetching alert rules")
	mockService.AssertExpectations(t)
}

func TestGetAlertRuleByIDHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	expectedRule := &alerting.AlertRule{
		ID:             1,
		SensorID:       1,
		SensorName:     "TestSensor",
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: 1,
		Enabled:        true,
	}

	mockService.On("ServiceGetAlertRuleByID", mock.Anything, 1).Return(expectedRule, nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/:id", s.getAlertRuleByIDHandler)

	req := httptest.NewRequest("GET", "/api/alerts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "TestSensor")
	mockService.AssertExpectations(t)
}

func TestGetAlertRuleByIDHandler_InvalidID(t *testing.T) {
	s := new(Server)
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/:id", s.getAlertRuleByIDHandler)

	req := httptest.NewRequest("GET", "/api/alerts/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule ID")
}

func TestGetAlertRuleByIDHandler_NotFound(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	mockService.On("ServiceGetAlertRuleByID", mock.Anything, 999).Return(nil, nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/:id", s.getAlertRuleByIDHandler)

	req := httptest.NewRequest("GET", "/api/alerts/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetAlertRulesBySensorIDHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	expectedRules := []alerting.AlertRule{
		{ID: 1, SensorID: 1, SensorName: "TestSensor", AlertType: alerting.AlertTypeNumericRange},
		{ID: 2, SensorID: 1, SensorName: "TestSensor", AlertType: alerting.AlertTypeStatusBased},
	}

	mockService.On("ServiceGetAlertRulesBySensorID", mock.Anything, 1).Return(expectedRules, nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/sensor/:sensorId", s.getAlertRulesBySensorIDHandler)

	req := httptest.NewRequest("GET", "/api/alerts/sensor/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "TestSensor")
	mockService.AssertExpectations(t)
}

func TestCreateAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	newRule := alerting.AlertRule{
		SensorID:          1,
		MeasurementTypeId: 1,
		AlertType:         alerting.AlertTypeNumericRange,
		HighThreshold:     30.0,
		LowThreshold:      10.0,
		RateLimitSeconds:    1,
		Enabled:           true,
	}

	mockService.On("ServiceCreateAlertRule", mock.Anything, mock.AnythingOfType("*alerting.AlertRule")).Return(nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	body, _ := json.Marshal(newRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule created successfully")
	mockService.AssertExpectations(t)
}

func TestCreateAlertRuleHandler_InvalidJSON(t *testing.T) {
	s := new(Server)
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestCreateAlertRuleHandler_ValidationError(t *testing.T) {
	s := new(Server)
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:      1,
		AlertType:     alerting.AlertTypeNumericRange,
		HighThreshold: 10.0, // Invalid: lower than low threshold
		LowThreshold:  30.0,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestCreateAlertRuleHandler_NegativeRateLimit(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}
	
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:       1,
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: -1, // Invalid: negative rate limit
		Enabled:        true,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestCreateAlertRuleHandler_ZeroSensorID(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}
	
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:       0, // Invalid: sensor ID must be positive
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: 1,
		Enabled:        true,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestCreateAlertRuleHandler_NegativeSensorID(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}
	
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:       -5, // Invalid: negative sensor ID
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: 1,
		Enabled:        true,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestCreateAlertRuleHandler_InvalidAlertType(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}
	
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.POST("/alerts", s.createAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		SensorID:       1,
		AlertType:      "invalid_type", // Invalid: not a recognized alert type
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: 1,
		Enabled:        true,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("POST", "/api/alerts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestUpdateAlertRuleHandler_NegativeRateLimit(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}
	
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.PUT("/alerts/:id", s.updateAlertRuleHandler)

	invalidRule := alerting.AlertRule{
		AlertType:      alerting.AlertTypeNumericRange,
		HighThreshold:  30.0,
		LowThreshold:   10.0,
		RateLimitSeconds: -1,
		Enabled:        true,
	}

	body, _ := json.Marshal(invalidRule)
	req := httptest.NewRequest("PUT", "/api/alerts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid alert rule")
}

func TestUpdateAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	updatedRule := alerting.AlertRule{
		SensorID:          1,
		MeasurementTypeId: 1,
		AlertType:         alerting.AlertTypeNumericRange,
		HighThreshold:     35.0,
		LowThreshold:      12.0,
		RateLimitSeconds:    2,
		Enabled:           false,
	}

	mockService.On("ServiceUpdateAlertRule", mock.Anything, mock.AnythingOfType("*alerting.AlertRule")).Return(nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.PUT("/alerts/:id", s.updateAlertRuleHandler)

	body, _ := json.Marshal(updatedRule)
	req := httptest.NewRequest("PUT", "/api/alerts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule updated successfully")
	mockService.AssertExpectations(t)
}

func TestDeleteAlertRuleHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	mockService.On("ServiceDeleteAlertRule", mock.Anything, 1).Return(nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.DELETE("/alerts/:id", s.deleteAlertRuleHandler)

	req := httptest.NewRequest("DELETE", "/api/alerts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Alert rule deleted successfully")
	mockService.AssertExpectations(t)
}

func TestGetAlertHistoryHandler(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	expectedHistory := []gen.AlertHistoryEntry{
		{SensorId: 1, AlertType: "numeric_range", ReadingValue: "35.5", SentAt: time.Now()},
		{SensorId: 1, AlertType: "numeric_range", ReadingValue: "40.0", SentAt: time.Now().Add(-2 * time.Hour)},
	}

	mockService.On("ServiceGetAlertHistory", mock.Anything, 1, 10).Return(expectedHistory, nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/sensor/:sensorId/history", s.getAlertHistoryHandler)

	req := httptest.NewRequest("GET", "/api/alerts/sensor/1/history?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "35.5")
	assert.Contains(t, w.Body.String(), "numeric_range")
	mockService.AssertExpectations(t)
}

func TestGetAlertHistoryHandler_DefaultLimit(t *testing.T) {
	mockService := new(mockAlertManagementService)
	s := &Server{alertService: mockService}

	mockService.On("ServiceGetAlertHistory", mock.Anything, 1, 50).Return([]gen.AlertHistoryEntry{}, nil)

	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.GET("/alerts/sensor/:sensorId/history", s.getAlertHistoryHandler)

	req := httptest.NewRequest("GET", "/api/alerts/sensor/1/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
