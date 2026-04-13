package api

import (
	"context"
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
	apiGroup := router.Group("/api")
	apiGroup.GET(route, handler)
	return router
}

type mockReadingsService struct {
	ServiceGetBetweenDatesFunc func(context.Context, string, string, string, string, bool) ([]types.Reading, error)
	ServiceGetLatestFunc       func(context.Context) ([]types.Reading, error)
}

func (m *mockReadingsService) ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
	return m.ServiceGetBetweenDatesFunc(ctx, startDate, endDate, sensorName, measurementType, hourly)
}
func (m *mockReadingsService) ServiceGetLatest(ctx context.Context) ([]types.Reading, error) {
	return m.ServiceGetLatestFunc(ctx)
}

func mockGetLatestReadingsSuccessful() ([]types.Reading, error) {
	val := 21.5
	return []types.Reading{
		{SensorName: "Test", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2025-08-31T10:00:00Z"},
	}, nil
}

func mockGetReadingsBetweenDatesSuccessful(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
	v1 := 22.5
	v2 := 24.5
	v3 := 23.5
	readings := []types.Reading{
		{SensorName: "sensor1", MeasurementType: "temperature", Unit: "°C", NumericValue: &v1, Time: "2024-01-01T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v1, Time: "2024-01-02T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v2, Time: "2024-01-03T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v3, Time: "2024-01-04T10:00:00Z"},
	}
	return readings, nil
}

func mockGetReadingsBetweenDatesError(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
	return nil, fmt.Errorf("failed to fetch readings")
}

func TestSuccessfulGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesSuccessful,
	}

	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingStartDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/hourly/between?end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingEndDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/hourly/between?start=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidStartDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/api/readings/hourly/between?start=invalid-date&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid start parameter")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidEndDate(t *testing.T) {
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/api/readings/hourly/between?start=2024-01-01&end=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid end parameter")
}

func TestErrorGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesError,
	}

	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to fetch readings")
}

func TestSuccessfulGetReadingsBetweenDatesHandler(t *testing.T) {
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesSuccessful,
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestGetReadingsBetweenDatesHandler_WithSensorFilter(t *testing.T) {
	var capturedSensor string
	val := 21.0
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
			capturedSensor = sensorName
			return []types.Reading{
				{SensorName: "Office", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2024-01-01T10:00:00Z"},
			}, nil
		},
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04&sensor=Office", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Office", capturedSensor)
	assert.Contains(t, w.Body.String(), "Office")
}

func TestGetReadingsBetweenDatesHandler_WithoutSensorFilter(t *testing.T) {
	var capturedSensor string
	val := 22.5
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
			capturedSensor = sensorName
			return []types.Reading{
				{SensorName: "sensor1", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2024-01-01T10:00:00Z"},
			}, nil
		},
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", capturedSensor)
}

func TestGetReadingsBetweenDatesHandler_ISODatetime(t *testing.T) {
	var capturedStart, capturedEnd string
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return []types.Reading{}, nil
		},
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01T10:00:00Z&end=2024-01-01T16:00:00Z", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 10:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 16:00:00", capturedEnd)
}

func TestGetReadingsBetweenDatesHandler_ISODatetimeWithOffset(t *testing.T) {
	var capturedStart, capturedEnd string
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return []types.Reading{}, nil
		},
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01T11:00:00%2B01:00&end=2024-01-01T17:00:00%2B01:00", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 10:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 16:00:00", capturedEnd)
}

func TestGetReadingsBetweenDatesHandler_DateOnlyExpandsToFullDay(t *testing.T) {
	var capturedStart, capturedEnd string
	readingsService = &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, hourly bool) ([]types.Reading, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return []types.Reading{}, nil
		},
	}

	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 00:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 23:59:59", capturedEnd)
}
