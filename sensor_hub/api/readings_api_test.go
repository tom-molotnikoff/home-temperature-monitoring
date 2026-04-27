package api

import (
	"context"
	gen "example/sensorHub/gen"
	
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
	ServiceGetBetweenDatesFunc func(context.Context, string, string, string, string, string, string) (*gen.AggregatedReadingsResponse, error)
	ServiceGetLatestFunc       func(context.Context) ([]gen.Reading, error)
}

func (m *mockReadingsService) ServiceGetBetweenDates(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
	return m.ServiceGetBetweenDatesFunc(ctx, startDate, endDate, sensorName, measurementType, overrideInterval, overrideFunction)
}
func (m *mockReadingsService) ServiceGetLatest(ctx context.Context) ([]gen.Reading, error) {
	return m.ServiceGetLatestFunc(ctx)
}

func mockGetLatestReadingsSuccessful() ([]gen.Reading, error) {
	val := 21.5
	return []gen.Reading{
		{SensorName: "Test", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2025-08-31T10:00:00Z"},
	}, nil
}

func mockGetReadingsBetweenDatesSuccessful(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
	v1 := 22.5
	v2 := 24.5
	v3 := 23.5
	readings := []gen.Reading{
		{SensorName: "sensor1", MeasurementType: "temperature", Unit: "°C", NumericValue: &v1, Time: "2024-01-01T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v1, Time: "2024-01-02T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v2, Time: "2024-01-03T10:00:00Z"},
		{SensorName: "sensor2", MeasurementType: "temperature", Unit: "°C", NumericValue: &v3, Time: "2024-01-04T10:00:00Z"},
	}
	return &gen.AggregatedReadingsResponse{
		AggregationInterval: "PT1H",
		AggregationFunction: "avg",
		Readings:            readings,
	}, nil
}

func mockGetReadingsBetweenDatesError(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
	return nil, fmt.Errorf("failed to fetch readings")
}

func TestSuccessfulGetReadingsBetweenDatesHandler(t *testing.T) {
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesSuccessful,
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
	assert.Contains(t, w.Body.String(), "aggregation_interval")
	assert.Contains(t, w.Body.String(), "aggregation_function")
}

func TestGetReadingsBetweenDatesHandler_MissingStartDate(t *testing.T) {
	s := new(Server)

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetReadingsBetweenDatesHandler_MissingEndDate(t *testing.T) {
	s := new(Server)

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetReadingsBetweenDatesHandler_InvalidStartDate(t *testing.T) {
	s := new(Server)

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/api/readings/between?start=invalid-date&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid start parameter")
}

func TestGetReadingsBetweenDatesHandler_InvalidEndDate(t *testing.T) {
	s := new(Server)

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)
	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid end parameter")
}

func TestErrorGetReadingsBetweenDatesHandler(t *testing.T) {
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: mockGetReadingsBetweenDatesError,
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to fetch readings")
}

func TestGetReadingsBetweenDatesHandler_WithSensorFilter(t *testing.T) {
	var capturedSensor string
	val := 21.0
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedSensor = sensorName
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: "raw",
				AggregationFunction: "none",
				Readings: []gen.Reading{
					{SensorName: "Office", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2024-01-01T10:00:00Z"},
				},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

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
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedSensor = sensorName
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: "raw",
				AggregationFunction: "none",
				Readings: []gen.Reading{
					{SensorName: "sensor1", MeasurementType: "temperature", Unit: "°C", NumericValue: &val, Time: "2024-01-01T10:00:00Z"},
				},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", capturedSensor)
}

func TestGetReadingsBetweenDatesHandler_ISODatetime(t *testing.T) {
	var capturedStart, capturedEnd string
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: "raw",
				AggregationFunction: "none",
				Readings:            []gen.Reading{},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01T10:00:00Z&end=2024-01-01T16:00:00Z", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 10:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 16:00:00", capturedEnd)
}

func TestGetReadingsBetweenDatesHandler_ISODatetimeWithOffset(t *testing.T) {
	var capturedStart, capturedEnd string
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: "raw",
				AggregationFunction: "none",
				Readings:            []gen.Reading{},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01T11:00:00%2B01:00&end=2024-01-01T17:00:00%2B01:00", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 10:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 16:00:00", capturedEnd)
}

func TestGetReadingsBetweenDatesHandler_DateOnlyExpandsToFullDay(t *testing.T) {
	var capturedStart, capturedEnd string
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedStart = startDate
			capturedEnd = endDate
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: "raw",
				AggregationFunction: "none",
				Readings:            []gen.Reading{},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "2024-01-01 00:00:00", capturedStart)
	assert.Equal(t, "2024-01-01 23:59:59", capturedEnd)
}

func TestGetReadingsBetweenDatesHandler_AggregationOverrideParams(t *testing.T) {
	var capturedInterval, capturedFunction string
	s := &Server{readingsService: &mockReadingsService{
		ServiceGetBetweenDatesFunc: func(ctx context.Context, startDate, endDate, sensorName, measurementType string, overrideInterval, overrideFunction string) (*gen.AggregatedReadingsResponse, error) {
			capturedInterval = overrideInterval
			capturedFunction = overrideFunction
			return &gen.AggregatedReadingsResponse{
				AggregationInterval: gen.AggregatedReadingsResponseAggregationInterval(overrideInterval),
				AggregationFunction: gen.AggregatedReadingsResponseAggregationFunction(overrideFunction),
				Readings:            []gen.Reading{},
			}, nil
		},
	}}

	router := setupTestRouter("/readings/between", s.getReadingsBetweenDatesHandler)

	req := httptest.NewRequest("GET", "/api/readings/between?start=2024-01-01&end=2024-01-04&aggregation=PT1H&aggregation_function=count", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "PT1H", capturedInterval)
	assert.Equal(t, "count", capturedFunction)
}
