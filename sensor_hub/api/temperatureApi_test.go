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
	ServiceGetBetweenDatesFunc func(string, string, string) ([]types.TemperatureReading, error)
	ServiceGetLatestFunc       func() ([]types.TemperatureReading, error)
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
