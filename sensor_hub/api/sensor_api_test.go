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
)

func init() {
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		HealthHistoryDefaultResponseNumber: 10,
	}
}

func setupSensorRouter() (*gin.Engine, *MockSensorService) {
	mockService := new(MockSensorService)
	InitSensorAPI(mockService)
	router := gin.New()
	return router, mockService
}

func TestAddSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors", addSensorHandler)

	sensor := types.Sensor{Name: "test-sensor", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", sensor).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetAllSensorsHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors", getAllSensorsHandler)

	mockService.On("ServiceGetAllSensors").Return([]types.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestGetSensorByNameHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", "s1").Return(&types.Sensor{Name: "s1"}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestUpdateSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)
	
	expectedSensor := sensor
	expectedSensor.Id = 1

	mockService.On("ServiceUpdateSensorById", expectedSensor).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.DELETE("/sensors/:name", deleteSensorHandler)

	mockService.On("ServiceDeleteSensorByName", "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/collect", collectAndStoreAllSensorReadingsHandler)

	mockService.On("ServiceCollectAndStoreAllSensorReadings").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCollectFromSensorByNameHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/collect", collectFromSensorByNameHandler)

	mockService.On("ServiceCollectFromSensorByName", "s1").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnableSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/enable", enableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", true).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDisableSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/disable", disableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", false).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTotalReadingsPerSensorHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/readings/total", totalReadingsPerSensorHandler)

	mockService.On("ServiceGetTotalReadingsForEachSensor").Return(map[string]int{"s1": 10}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "10")
}

func TestGetSensorsByTypeHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/type/:type", getSensorsByTypeHandler)

	mockService.On("ServiceGetSensorsByType", "Temperature").Return([]types.Sensor{{Name: "s1"}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/type/Temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "s1")
}

func TestSensorExistsHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", "s1").Return(true, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	mockService.On("ServiceGetSensorHealthHistoryByName", "s1", 10).Return([]types.SensorHealthHistory{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSensorHandler_InvalidJSON(t *testing.T) {
	router, _ := setupSensorRouter()
	router.POST("/sensors", addSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors", addSensorHandler)

	sensor := types.Sensor{Name: "test-sensor", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)

	mockService.On("ServiceAddSensor", sensor).Return(errors.New("validation error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorByNameHandler_NotFound(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", "notfound").Return((*types.Sensor)(nil), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSensorByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name", getSensorByNameHandler)

	mockService.On("ServiceGetSensorByName", "s1").Return((*types.Sensor)(nil), errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateSensorHandler_InvalidID(t *testing.T) {
	router, _ := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/invalid", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_InvalidJSON(t *testing.T) {
	router, _ := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/1", bytes.NewBufferString("invalid"))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.PUT("/sensors/:id", updateSensorHandler)

	sensor := types.Sensor{Name: "s1-updated", Type: "Temperature", URL: "http://localhost:8080"}
	jsonBody, _ := json.Marshal(sensor)
	
	expectedSensor := sensor
	expectedSensor.Id = 1

	mockService.On("ServiceUpdateSensorById", expectedSensor).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/sensors/1", bytes.NewBuffer(jsonBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.DELETE("/sensors/:name", deleteSensorHandler)

	mockService.On("ServiceDeleteSensorByName", "s1").Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllSensorsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors", getAllSensorsHandler)

	mockService.On("ServiceGetAllSensors").Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorsByTypeHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/type/:type", getSensorsByTypeHandler)

	mockService.On("ServiceGetSensorsByType", "Temperature").Return([]types.Sensor{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/type/Temperature", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSensorExistsHandler_NotFound(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", "notfound").Return(false, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/sensors/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSensorExistsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.HEAD("/sensors/:name", sensorExistsHandler)

	mockService.On("ServiceSensorExists", "s1").Return(false, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/sensors/s1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectAndStoreAllSensorReadingsHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/collect", collectAndStoreAllSensorReadingsHandler)

	mockService.On("ServiceCollectAndStoreAllSensorReadings").Return(errors.New("collection error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCollectFromSensorByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/collect", collectFromSensorByNameHandler)

	mockService.On("ServiceCollectFromSensorByName", "s1").Return(errors.New("sensor offline"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/collect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEnableSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/enable", enableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", true).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/enable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDisableSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.POST("/sensors/:sensorName/disable", disableSensorHandler)

	mockService.On("ServiceSetEnabledSensorByName", "s1", false).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/sensors/s1/disable", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTotalReadingsPerSensorHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/readings/total", totalReadingsPerSensorHandler)

	mockService.On("ServiceGetTotalReadingsForEachSensor").Return(map[string]int{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/readings/total", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_InvalidLimit(t *testing.T) {
	router, _ := setupSensorRouter()
	router.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1/health?limit=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSensorHealthHistoryByNameHandler_ServiceError(t *testing.T) {
	router, mockService := setupSensorRouter()
	router.GET("/sensors/:name/health", getSensorHealthHistoryByNameHandler)

	mockService.On("ServiceGetSensorHealthHistoryByName", "s1", 10).Return([]types.SensorHealthHistory{}, errors.New("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sensors/s1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
