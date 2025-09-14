package api

import (
	appProps "example/sensorHub/application_properties"
	database "example/sensorHub/db"
	"example/sensorHub/sensors"
	"example/sensorHub/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(route string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(route, handler)
	return router
}

func mockGetLatestReadingsSuccessful() ([]types.APIReading, error) {
	return []types.APIReading{
		{SensorName: "Test", Reading: types.TemperatureReading{Temperature: 21.5, Time: "2025-08-31T10:00:00Z"}},
	}, nil
}

func mockTakeReadingsSuccessful() ([]types.RawSensorReading, error) {
	return []types.RawSensorReading{
		{SensorName: "sensor1", Reading: types.TemperatureReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.TemperatureReading{Temperature: 23.0, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor3", Reading: types.TemperatureReading{Temperature: 21.5, Time: "2024-01-01T10:00:00Z"}},
	}, nil
}

func mockTakeReadingsError() ([]types.RawSensorReading, error) {
	return nil, fmt.Errorf("failed to collect readings")
}

func mockTakeReadingFromNamedSensorSuccessful(sensorName string) (*types.RawSensorReading, error) {
	return &types.RawSensorReading{SensorName: "sensor1", Reading: types.TemperatureReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}}, nil
}

func mockTakeReadingFromNamedSensorError(sensorName string) (*types.RawSensorReading, error) {
	return nil, fmt.Errorf("Something went wrong")
}

func mockGetReadingsBetweenDatesSuccessful(tableName, startDate, endDate string) ([]types.APIReading, error) {
	readings := []types.APIReading{
		{SensorName: "sensor1", Reading: types.TemperatureReading{Temperature: 22.5, Time: "2024-01-01T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.TemperatureReading{Temperature: 22.5, Time: "2024-01-02T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.TemperatureReading{Temperature: 24.5, Time: "2024-01-03T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.TemperatureReading{Temperature: 23.5, Time: "2024-01-04T10:00:00Z"}},
	}
	return readings, nil
}

func mockGetReadingsBetweenDatesError(tableName, startDate, endDate string) ([]types.APIReading, error) {
	return nil, fmt.Errorf("failed to fetch readings")
}

func TestSuccessfulCollectAllSensorsHandler(t *testing.T) {
	// Mock the take_readings function
	originalTakeReadings := sensors.TakeReadingsFromAllSensors
	sensors.TakeReadingsFromAllSensors = mockTakeReadingsSuccessful
	// Ensure we restore the original function after the test
	defer func() { sensors.TakeReadingsFromAllSensors = originalTakeReadings }()

	// Set up Gin for testing
	router := setupTestRouter("/sensors/temperature", collectAllSensorsHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/sensors/temperature", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "sensor2")
	assert.Contains(t, w.Body.String(), "sensor3")
}

func TestErrorCollectAllSensorsHandler(t *testing.T) {
	// Mock the take_readings function with an error
	originalTakeReadings := sensors.TakeReadingsFromAllSensors
	sensors.TakeReadingsFromAllSensors = mockTakeReadingsError
	defer func() { sensors.TakeReadingsFromAllSensors = originalTakeReadings }()

	// Set up Gin for testing
	router := setupTestRouter("/sensors/temperature", collectAllSensorsHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/sensors/temperature", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Error collecting readings")
}

func TestSuccessfulCollectSpecificSensorHandler(t *testing.T) {
	// Mock the take_reading_from_named_sensor function
	originalTakeReadingFromNamedSensor := sensors.TakeReadingFromNamedSensor
	sensors.TakeReadingFromNamedSensor = mockTakeReadingFromNamedSensorSuccessful
	defer func() { sensors.TakeReadingFromNamedSensor = originalTakeReadingFromNamedSensor }()

	// Set up Gin for testing
	router := setupTestRouter("/sensors/temperature/:sensorName", collectSpecificSensorHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/sensors/temperature/sensor1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestErrorCollectSpecificSensorHandler(t *testing.T) {
	// Mock the take_reading_from_named_sensor function with an error
	originalTakeReadingFromNamedSensor := sensors.TakeReadingFromNamedSensor
	sensors.TakeReadingFromNamedSensor = mockTakeReadingFromNamedSensorError
	defer func() { sensors.TakeReadingFromNamedSensor = originalTakeReadingFromNamedSensor }()

	// Set up Gin for testing
	router := setupTestRouter("/sensors/temperature/:sensorName", collectSpecificSensorHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/sensors/temperature/sensor1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Something went wrong")
}

func TestSuccessfulGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	// Mock the getReadingsBetweenDates function
	originalGetReadingsBetweenDates := database.GetReadingsBetweenDates
	database.GetReadingsBetweenDates = mockGetReadingsBetweenDatesSuccessful
	defer func() { database.GetReadingsBetweenDates = originalGetReadingsBetweenDates }()

	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingStartDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	// Make web request without start date
	req := httptest.NewRequest("GET", "/readings/hourly/between?end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_MissingEndDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	// Make web request without end date
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidStartDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	// Make web request with invalid start date
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=invalid-date&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid start date format")
}

func TestGetHourlyReadingsBetweenDatesHandler_InvalidEndDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)
	// Make web request with invalid end date
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid end date format")
}

func TestErrorGetHourlyReadingsBetweenDatesHandler(t *testing.T) {
	// Mock the getReadingsBetweenDates function with an error
	originalGetReadingsBetweenDates := database.GetReadingsBetweenDates
	database.GetReadingsBetweenDates = mockGetReadingsBetweenDatesError
	defer func() { database.GetReadingsBetweenDates = originalGetReadingsBetweenDates }()

	// Set up Gin for testing
	router := setupTestRouter("/readings/hourly/between", getHourlyReadingsBetweenDatesHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/readings/hourly/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to fetch readings")
}

func TestSuccessfulGetReadingsBetweenDatesHandler(t *testing.T) {
	// Mock the getReadingsBetweenDates function
	originalGetReadingsBetweenDates := database.GetReadingsBetweenDates
	database.GetReadingsBetweenDates = mockGetReadingsBetweenDatesSuccessful
	defer func() { database.GetReadingsBetweenDates = originalGetReadingsBetweenDates }()

	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensor1")
	assert.Contains(t, w.Body.String(), "22.5")
}

func TestGetReadingsBetweenDatesHandler_MissingStartDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request without start date
	req := httptest.NewRequest("GET", "/readings/between?end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetReadingsBetweenDatesHandler_MissingEndDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request without end date
	req := httptest.NewRequest("GET", "/readings/between?start=2024-01-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Start and end dates are required")
}

func TestGetReadingsBetweenDatesHandler_InvalidStartDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request with invalid start date
	req := httptest.NewRequest("GET", "/readings/between?start=invalid-date&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid start date format")
}

func TestGetReadingsBetweenDatesHandler_InvalidEndDate(t *testing.T) {
	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request with invalid end date
	req := httptest.NewRequest("GET", "/readings/between?start=2024-01-01&end=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid end date format")
}

func TestErrorGetReadingsBetweenDatesHandler(t *testing.T) {
	// Mock the getReadingsBetweenDates function with an error
	originalGetReadingsBetweenDates := database.GetReadingsBetweenDates
	database.GetReadingsBetweenDates = mockGetReadingsBetweenDatesError
	defer func() { database.GetReadingsBetweenDates = originalGetReadingsBetweenDates }()

	// Set up Gin for testing
	router := setupTestRouter("/readings/between", getReadingsBetweenDatesHandler)

	// Make web request
	req := httptest.NewRequest("GET", "/readings/between?start=2024-01-01&end=2024-01-04", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to fetch readings")
}

func TestSuccessfulCurrentTemperaturesWebSocket(t *testing.T) {
	// Mock getLatestReadings to return predictable data
	originalGetLatestReadings := database.GetLatestReadings
	database.GetLatestReadings = mockGetLatestReadingsSuccessful
	defer func() { database.GetLatestReadings = originalGetLatestReadings }()

	if appProps.APPLICATION_PROPERTIES == nil {
		appProps.APPLICATION_PROPERTIES = make(map[string]string)
		// Shorten the interval to a tiny amount for quick testing
		appProps.APPLICATION_PROPERTIES["current.temperature.websocket.interval"] = "0.01"
	}

	// Set up Gin for testing
	router := setupTestRouter("/ws/current-temperatures", currentTemperaturesWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	wsURL := "ws" + server.URL[len("http"):] + "/ws/current-temperatures"

	// Connect as a WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Read a message (should be sent within a few seconds)
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg, err := ws.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(msg), "Test") // Check for expected sensor name

	ws.Close()
}
