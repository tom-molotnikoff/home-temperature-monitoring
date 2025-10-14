package api

import (
	"example/sensorHub/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(route string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(route, handler)
	return router
}

type mockTemperatureService struct {
	ServiceCollectAllSensorReadingsFunc func() ([]types.APITempReading, error)
	ServiceCollectSensorReadingFunc     func(string) (*types.APITempReading, error)
	ServiceGetBetweenDatesFunc          func(string, string, string) ([]types.APITempReading, error)
	ServiceGetLatestFunc                func() ([]types.APITempReading, error)
}

func (m *mockTemperatureService) ServiceCollectAllSensorReadings() ([]types.APITempReading, error) {
	return m.ServiceCollectAllSensorReadingsFunc()
}
func (m *mockTemperatureService) ServiceCollectSensorReading(sensorName string) (*types.APITempReading, error) {
	return m.ServiceCollectSensorReadingFunc(sensorName)
}
func (m *mockTemperatureService) ServiceGetBetweenDates(table, start, end string) ([]types.APITempReading, error) {
	return m.ServiceGetBetweenDatesFunc(table, start, end)
}
func (m *mockTemperatureService) ServiceGetLatest() ([]types.APITempReading, error) {
	return m.ServiceGetLatestFunc()
}

func mockGetLatestReadingsSuccessful() ([]types.APITempReading, error) {
	return []types.APITempReading{
		{SensorName: "Test", Reading: types.RawTempReading{Temperature: 21.5, Time: "2025-08-31T10:00:00Z"}},
	}, nil
}

func mockGetReadingFromAllTemperatureSensorsSuccessful() ([]types.APITempReading, error) {
	return []types.APITempReading{
		{SensorName: "sensor1", Reading: types.RawTempReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.RawTempReading{Temperature: 23.0, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor3", Reading: types.RawTempReading{Temperature: 21.5, Time: "2024-01-01T10:00:00Z"}},
	}, nil
}

func mockGetReadingFromAllTemperatureSensorsError() ([]types.APITempReading, error) {
	return nil, fmt.Errorf("failed to collect readings")
}

func mockGetReadingFromTemperatureSensorSuccessful(sensorName string) (*types.APITempReading, error) {
	return &types.APITempReading{SensorName: "sensor1", Reading: types.RawTempReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}}, nil
}

func mockGetReadingFromTemperatureSensorError(sensorName string) (*types.APITempReading, error) {
	return nil, fmt.Errorf("Something went wrong")
}

func mockGetReadingsBetweenDatesSuccessful(tableName, startDate, endDate string) ([]types.APITempReading, error) {
	readings := []types.APITempReading{
		{SensorName: "sensor1", Reading: types.RawTempReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.RawTempReading{Temperature: 22.5, Time: "2024-01-02T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.RawTempReading{Temperature: 24.5, Time: "2024-01-03T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.RawTempReading{Temperature: 23.5, Time: "2024-01-04T10:00:00Z"}},
	}
	return readings, nil
}

func mockGetReadingsBetweenDatesError(tableName, startDate, endDate string) ([]types.APITempReading, error) {
	return nil, fmt.Errorf("failed to fetch readings")
}

func TestSuccessfulCollectAllSensorsHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceCollectAllSensorReadingsFunc: mockGetReadingFromAllTemperatureSensorsSuccessful,
	}

	router := setupTestRouter("/sensors/temperature", collectAllTemperatureSensorsHandler)

	req := httptest.NewRequest("GET", "/sensors/temperature", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "sensor2")
	assert.Contains(t, w.Body.String(), "sensor3")
}

func TestErrorCollectAllSensorsHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceCollectAllSensorReadingsFunc: mockGetReadingFromAllTemperatureSensorsError,
	}

	router := setupTestRouter("/sensors/temperature", collectAllTemperatureSensorsHandler)

	req := httptest.NewRequest("GET", "/sensors/temperature", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Error collecting readings")
}

func TestSuccessfulCollectSpecificSensorHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceCollectSensorReadingFunc: mockGetReadingFromTemperatureSensorSuccessful,
	}

	router := setupTestRouter("/sensors/temperature/:sensorName", collectSpecificTemperatureSensorHandler)

	req := httptest.NewRequest("GET", "/sensors/temperature/sensor1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestErrorCollectSpecificSensorHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceCollectSensorReadingFunc: mockGetReadingFromTemperatureSensorError,
	}

	router := setupTestRouter("/sensors/temperature/:sensorName", collectSpecificTemperatureSensorHandler)

	req := httptest.NewRequest("GET", "/sensors/temperature/sensor1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Something went wrong")
}

func TestSuccessfulGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesSuccessful,
	}

	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingStartDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/readings/hourly/between?end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingEndDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidStartDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=invalid-date&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid start date format")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidEndDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid end date format")
}

func TestErrorGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesError,
	}

	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to fetch readings")
}

func TestSuccessfulGetReadingsBetweenDatesHandler(t *testing.T) {
	tempService = &mockTemperatureService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesSuccessful,
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}
