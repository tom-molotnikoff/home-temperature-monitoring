package api

import (
	"bytes"
	"encoding/json"
	"errors"
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
func init() {
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		HealthHistoryDefaultResponseNumber: 10,
	}
}

func setupSensorRouter() (*gin.Engine, *gin.RouterGroup, *MockSensorService) {
	mockService := new(MockSensorService)
	InitSensorAPI(mockService)
	router := gin.New()
	apiGroup := router.Group("/api")
	return router, apiGroup, mockService
}

func TestAddSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors", addSensorHandler)

	sensor := types.Sensor{Name: "test-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", mock.Anything, sensor).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetAllSensorsHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors", getAllSensorsHandler)

	mockService.On("ServiceGetAllSensors", mock.Anything).Return([]types.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestGetSensorByNameHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", mock.Anything, "s1").Return(&types.Sensor{Name: "s1"}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestUpdateSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)
	
	expectedSensor := sensor
	expectedSensor.Id = 1

	mockService.On("ServiceUpdateSensorById", mock.Anything, expectedSensor).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.DELETE("/sensors/:name", deleteSensorHandler)

	mockService.On("ServiceDeleteSensorByName", mock.Anything, "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/collect", collectAndStoreAllSensorReadingsHandler)

	mockService.On("ServiceCollectAndStoreAllSensorReadings", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectFromSensorByNameHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/collect", collectFromSensorByNameHandler)

	mockService.On("ServiceCollectFromSensorByName", mock.Anything, "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnableSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/enable", enableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", true).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDisableSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/disable", disableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", false).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTotalReadingsPerSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/readings/total", totalReadingsPerSensorHandler)

	mockService.On("ServiceGetTotalReadingsForEachSensor", mock.Anything).Return(map[string]int{"s1": 10}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "10")
}

func TestGetSensorsByDriverHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/driver/:driver", getSensorsByDriverHandler)

	mockService.On("ServiceGetSensorsByDriver", mock.Anything, "sensor-hub-http-temperature").Return([]types.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/driver/sensor-hub-http-temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestSensorExistsHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", mock.Anything, "s1").Return(true, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	mockService.On("ServiceGetSensorHealthHistoryByName", mock.Anything, "s1", 10).Return([]types.SensorHealthHistory{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSensorHandler_InvalidJSON(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.POST("/sensors", addSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors", addSensorHandler)

	sensor := types.Sensor{Name: "test-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", mock.Anything, sensor).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorByNameHandler_NotFound(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", mock.Anything, "notfound").Return((*types.Sensor)(nil), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSensorByNameHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", mock.Anything, "s1").Return((*types.Sensor)(nil), errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateSensorHandler_InvalidID(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/invalid", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_InvalidJSON(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.PUT("/sensors/:id", updateSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}
	jsonBody, _ := json.Marshal(sensor)
	
	expectedSensor := sensor
	expectedSensor.Id = 1

	mockService.On("ServiceUpdateSensorById", mock.Anything, expectedSensor).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.DELETE("/sensors/:name", deleteSensorHandler)

	mockService.On("ServiceDeleteSensorByName", mock.Anything, "s1").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllSensorsHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors", getAllSensorsHandler)

	mockService.On("ServiceGetAllSensors", mock.Anything).Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorsByDriverHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/driver/:driver", getSensorsByDriverHandler)

	mockService.On("ServiceGetSensorsByDriver", mock.Anything, "sensor-hub-http-temperature").Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/driver/sensor-hub-http-temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSensorExistsHandler_NotFound(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", mock.Anything, "notfound").Return(false, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSensorExistsHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", mock.Anything, "s1").Return(false, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/api/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/collect", collectAndStoreAllSensorReadingsHandler)

	mockService.On("ServiceCollectAndStoreAllSensorReadings", mock.Anything).Return(errors.New("collection error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectFromSensorByNameHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/collect", collectFromSensorByNameHandler)

	mockService.On("ServiceCollectFromSensorByName", mock.Anything, "s1").Return(errors.New("sensor offline"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEnableSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/enable", enableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDisableSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/:sensorName/disable", disableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", mock.Anything, "s1", false).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTotalReadingsPerSensorHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/readings/total", totalReadingsPerSensorHandler)

	mockService.On("ServiceGetTotalReadingsForEachSensor", mock.Anything).Return(map[string]int{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_InvalidLimit(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health?limit=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_ServiceError(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	mockService.On("ServiceGetSensorHealthHistoryByName", mock.Anything, "s1", 10).Return([]types.SensorHealthHistory{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================================
// Sensor Status Handlers
// ============================================================================

func TestGetSensorsByStatusHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", getSensorsByStatusHandler)

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return([]types.Sensor{
		{Id: 1, Name: "auto-sensor", Status: "pending"},
	}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "auto-sensor")
}

func TestGetSensorsByStatusHandler_Empty(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", getSensorsByStatusHandler)

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return(nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[]")
}

func TestGetSensorsByStatusHandler_Error(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.GET("/sensors/status/:status", getSensorsByStatusHandler)

	mockService.On("ServiceGetSensorsByStatus", mock.Anything, "pending").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/sensors/status/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestApproveSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/approve/:id", approveSensorHandler)

	mockService.On("ServiceApproveSensor", mock.Anything, 1).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "approved")
}

func TestApproveSensorHandler_InvalidID(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.POST("/sensors/approve/:id", approveSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApproveSensorHandler_Error(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/approve/:id", approveSensorHandler)

	mockService.On("ServiceApproveSensor", mock.Anything, 1).Return(errors.New("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/approve/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDismissSensorHandler(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", dismissSensorHandler)

	mockService.On("ServiceDismissSensor", mock.Anything, 1).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "dismissed")
}

func TestDismissSensorHandler_InvalidID(t *testing.T) {
	router, api, _ := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", dismissSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDismissSensorHandler_Error(t *testing.T) {
	router, api, mockService := setupSensorRouter()
	api.POST("/sensors/dismiss/:id", dismissSensorHandler)

	mockService.On("ServiceDismissSensor", mock.Anything, 1).Return(errors.New("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/sensors/dismiss/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
