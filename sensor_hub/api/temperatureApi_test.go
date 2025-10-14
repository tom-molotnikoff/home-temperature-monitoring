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
	ServiceCollectAllSensorReadingsFunc func() ([]types.TemperatureReading, error)
	ServiceCollectSensorReadingFunc     func(string) (*types.TemperatureReading, error)
	ServiceGetBetweenDatesFunc          func(string, string, string) ([]types.TemperatureReading, error)
	ServiceGetLatestFunc                func() ([]types.TemperatureReading, error)
}

func (m *mockTemperatureService) ServiceCollectAllSensorReadings() ([]types.TemperatureReading, error) {
	return m.ServiceCollectAllSensorReadingsFunc()
}
func (m *mockTemperatureService) ServiceCollectSensorReading(sensorName string) (*types.TemperatureReading, error) {
	return m.ServiceCollectSensorReadingFunc(sensorName)
}
func (m *mockTemperatureService) ServiceGetBetweenDates(table, start, end string) ([]types.TemperatureReading, error) {
	return m.ServiceGetBetweenDatesFunc(table, start, end)
}
func (m *mockTemperatureService) ServiceGetLatest() ([]types.TemperatureReading, error) {
	return m.ServiceGetLatestFunc()
}

func mockGetLatestReadingsSuccessful() ([]types.TemperatureReading, error) {
	return []types.TemperatureReading{
		{SensorName: "Test", Temperature: 21.5, Time: "2025-08-31T10:00:00Z"},
	}, nil
}

func mockGetReadingFromAllTemperatureSensorsSuccessful() ([]types.TemperatureReading, error) {
	return []types.TemperatureReading{
		{SensorName: "sensor1", Temperature: 22.5, Time: "2024-01-01T10:00:00Z"},
		{SensorName: "sensor2", Temperature: 23.0, Time: "2024-01-01T10:00:00Z"},
		{SensorName: "sensor3", Temperature: 21.5, Time: "2024-01-01T10:00:00Z"},
	}, nil
}

func mockGetReadingFromAllTemperatureSensorsError() ([]types.TemperatureReading, error) {
	return nil, fmt.Errorf("failed to collect readings")
}

func mockGetReadingFromTemperatureSensorSuccessful(sensorName string) (*types.TemperatureReading, error) {
	return &types.TemperatureReading{SensorName: "sensor1", Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}, nil
}

func mockGetReadingFromTemperatureSensorError(sensorName string) (*types.TemperatureReading, error) {
	return nil, fmt.Errorf("Something went wrong")
}

func mockGetReadingsBetweenDatesSuccessful(tableName, startDate, endDate string) ([]types.TemperatureReading, error) {
	readings := []types.TemperatureReading{
		{SensorName: "sensor1", Temperature: 22.5, Time: "2024-01-01T10:00:00Z"},
		{SensorName: "sensor2", Temperature: 22.5, Time: "2024-01-02T10:00:00Z"},
		{SensorName: "sensor2", Temperature: 24.5, Time: "2024-01-03T10:00:00Z"},
		{SensorName: "sensor2", Temperature: 23.5, Time: "2024-01-04T10:00:00Z"},
	}
	return readings, nil
}

func mockGetReadingsBetweenDatesError(tableName, startDate, endDate string) ([]types.TemperatureReading, error) {
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
