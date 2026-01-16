package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example/sensorHub/alerting"
	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock AlertRepository for SensorService
// ============================================================================

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) GetAlertRuleBySensorID(sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) UpdateLastAlertSent(ruleID int) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) RecordAlertSent(ruleID, sensorID int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ruleID, sensorID, reason, numericValue, statusValue)
	return args.Error(0)
}

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

// Mock Notifier
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendAlert(sensorName, sensorType, reason string, numericValue float64, statusValue string) error {
	args := m.Called(sensorName, sensorType, reason, numericValue, statusValue)
	return args.Error(0)
}

// ============================================================================
// Test helpers
// ============================================================================

func setupSensorService() (*SensorService, *MockSensorRepository, *MockTemperatureRepository, *MockAlertRepository) {
	sensorRepo := new(MockSensorRepository)
	tempRepo := new(MockTemperatureRepository)
	alertRepo := new(MockAlertRepository)
	notifier := new(MockNotifier)

	service := NewSensorService(sensorRepo, tempRepo, alertRepo, notifier)
	return service, sensorRepo, tempRepo, alertRepo
}

// ============================================================================
// ServiceAddSensor tests
// ============================================================================

func TestSensorService_ServiceAddSensor_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", Type: "Temperature", URL: "http://localhost:8080", Enabled: true}

	// Create a mock HTTP server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.URL = server.URL

	sensorRepo.On("SensorExists", "TestSensor").Return(false, nil)
	sensorRepo.On("AddSensor", mock.Anything).Return(nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceAddSensor(sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond) // Allow async goroutine to complete
}

func TestSensorService_ServiceAddSensor_AlreadyExists(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "ExistingSensor", Type: "Temperature", URL: "http://localhost:8080"}

	// Create mock server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.URL = server.URL

	sensorRepo.On("SensorExists", "ExistingSensor").Return(true, nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceAddSensor(sensor)

	assert.Error(t, err)
	var alreadyExistsErr *AlreadyExistsError
	assert.True(t, errors.As(err, &alreadyExistsErr))
}

func TestSensorService_ServiceAddSensor_ValidationError_EmptyName(t *testing.T) {
	service, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "", Type: "Temperature", URL: "http://localhost:8080"}

	err := service.ServiceAddSensor(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor validation failed")
}

func TestSensorService_ServiceAddSensor_SensorExistsError(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", Type: "Temperature", URL: "http://localhost:8080"}

	// Create mock server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.URL = server.URL

	sensorRepo.On("SensorExists", "TestSensor").Return(false, errors.New("database error"))
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceAddSensor(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error checking if sensor exists")
}

// ============================================================================
// ServiceUpdateSensorById tests
// ============================================================================

func TestSensorService_ServiceUpdateSensorById_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "UpdatedSensor", Type: "Temperature", URL: "http://localhost:8080"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.URL = server.URL

	sensorRepo.On("UpdateSensorById", mock.Anything).Return(nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceUpdateSensorById(sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceUpdateSensorById_ValidationError(t *testing.T) {
	service, _, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "", Type: "Temperature", URL: "http://localhost:8080"}

	err := service.ServiceUpdateSensorById(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor validation failed")
}

// ============================================================================
// ServiceDeleteSensorByName tests
// ============================================================================

func TestSensorService_ServiceDeleteSensorByName_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "TestSensor").Return(true, nil)
	sensorRepo.On("DeleteSensorByName", "TestSensor").Return(nil)
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceDeleteSensorByName("TestSensor")

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceDeleteSensorByName_NotExists(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "NonExistent").Return(false, nil)

	err := service.ServiceDeleteSensorByName("NonExistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestSensorService_ServiceDeleteSensorByName_Error(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "TestSensor").Return(true, nil)
	sensorRepo.On("DeleteSensorByName", "TestSensor").Return(errors.New("database error"))

	err := service.ServiceDeleteSensorByName("TestSensor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting sensor")
}

// ============================================================================
// ServiceGetSensorByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorByName_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensor := &types.Sensor{Id: 1, Name: "TestSensor", Type: "Temperature"}
	sensorRepo.On("GetSensorByName", "TestSensor").Return(sensor, nil)

	result, err := service.ServiceGetSensorByName("TestSensor")

	assert.NoError(t, err)
	assert.Equal(t, "TestSensor", result.Name)
}

func TestSensorService_ServiceGetSensorByName_EmptyName(t *testing.T) {
	service, _, _, _ := setupSensorService()

	result, err := service.ServiceGetSensorByName("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor name cannot be empty")
	assert.Nil(t, result)
}

func TestSensorService_ServiceGetSensorByName_NotFound(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetSensorByName", "NonExistent").Return(nil, nil)

	result, err := service.ServiceGetSensorByName("NonExistent")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// ServiceGetAllSensors tests
// ============================================================================

func TestSensorService_ServiceGetAllSensors_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "Sensor1"},
		{Id: 2, Name: "Sensor2"},
	}
	sensorRepo.On("GetAllSensors").Return(sensors, nil)

	result, err := service.ServiceGetAllSensors()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestSensorService_ServiceGetAllSensors_Empty(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, nil)

	result, err := service.ServiceGetAllSensors()

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestSensorService_ServiceGetAllSensors_Error(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, errors.New("database error"))

	_, err := service.ServiceGetAllSensors()

	assert.Error(t, err)
}

// ============================================================================
// ServiceGetSensorsByType tests
// ============================================================================

func TestSensorService_ServiceGetSensorsByType_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "TempSensor1", Type: "Temperature"},
		{Id: 2, Name: "TempSensor2", Type: "Temperature"},
	}
	sensorRepo.On("GetSensorsByType", "Temperature").Return(sensors, nil)

	result, err := service.ServiceGetSensorsByType("Temperature")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// ============================================================================
// ServiceGetSensorIdByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorIdByName_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", "TestSensor").Return(1, nil)

	result, err := service.ServiceGetSensorIdByName("TestSensor")

	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestSensorService_ServiceGetSensorIdByName_Error(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", "NonExistent").Return(0, errors.New("not found"))

	_, err := service.ServiceGetSensorIdByName("NonExistent")

	assert.Error(t, err)
}

// ============================================================================
// ServiceSensorExists tests
// ============================================================================

func TestSensorService_ServiceSensorExists_True(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "ExistingSensor").Return(true, nil)

	result, err := service.ServiceSensorExists("ExistingSensor")

	assert.NoError(t, err)
	assert.True(t, result)
}

func TestSensorService_ServiceSensorExists_False(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "NonExistent").Return(false, nil)

	result, err := service.ServiceSensorExists("NonExistent")

	assert.NoError(t, err)
	assert.False(t, result)
}

// ============================================================================
// ServiceSetEnabledSensorByName tests
// ============================================================================

func TestSensorService_ServiceSetEnabledSensorByName_Enable(t *testing.T) {
	service, sensorRepo, tempRepo, alertRepo := setupSensorService()

	sensor := &types.Sensor{Id: 1, Name: "TestSensor", Type: "Temperature", Enabled: true}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.URL = server.URL

	sensorRepo.On("SensorExists", "TestSensor").Return(true, nil)
	sensorRepo.On("SetEnabledSensorByName", "TestSensor", true).Return(nil)
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{*sensor}, nil).Maybe()
	sensorRepo.On("GetSensorByName", "TestSensor").Return(sensor, nil).Maybe()
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	tempRepo.On("Add", mock.Anything).Return(nil).Maybe()
	// The async collection triggers alert processing
	alertRepo.On("GetAlertRuleBySensorID", mock.Anything).Return(nil, nil).Maybe()

	err := service.ServiceSetEnabledSensorByName("TestSensor", true)

	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
}

func TestSensorService_ServiceSetEnabledSensorByName_Disable(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "TestSensor").Return(true, nil)
	sensorRepo.On("SetEnabledSensorByName", "TestSensor", false).Return(nil)
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceSetEnabledSensorByName("TestSensor", false)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceSetEnabledSensorByName_NotExists(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", "NonExistent").Return(false, nil)

	err := service.ServiceSetEnabledSensorByName("NonExistent", true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// ============================================================================
// ServiceGetTotalReadingsForEachSensor tests
// ============================================================================

func TestSensorService_ServiceGetTotalReadingsForEachSensor_Success(t *testing.T) {
	service, sensorRepo, tempRepo, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "Sensor1", Type: "Temperature"},
		{Id: 2, Name: "Sensor2", Type: "Temperature"},
	}
	sensorRepo.On("GetAllSensors").Return(sensors, nil)
	tempRepo.On("GetTotalReadingsBySensorId", 1).Return(100, nil)
	tempRepo.On("GetTotalReadingsBySensorId", 2).Return(50, nil)

	result, err := service.ServiceGetTotalReadingsForEachSensor()

	assert.NoError(t, err)
	assert.Equal(t, 100, result["Sensor1"])
	assert.Equal(t, 50, result["Sensor2"])
}

func TestSensorService_ServiceGetTotalReadingsForEachSensor_Error(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetAllSensors").Return([]types.Sensor{}, errors.New("database error"))

	_, err := service.ServiceGetTotalReadingsForEachSensor()

	assert.Error(t, err)
}

// ============================================================================
// ServiceGetSensorHealthHistoryByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorHealthHistoryByName_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	history := []types.SensorHealthHistory{
		{SensorId: "1", HealthStatus: types.SensorGoodHealth},
	}
	sensorRepo.On("GetSensorIdByName", "TestSensor").Return(1, nil)
	sensorRepo.On("GetSensorHealthHistoryById", 1, 10).Return(history, nil)

	result, err := service.ServiceGetSensorHealthHistoryByName("TestSensor", 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSensorService_ServiceGetSensorHealthHistoryByName_SensorNotFound(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", "NonExistent").Return(0, errors.New("not found"))

	_, err := service.ServiceGetSensorHealthHistoryByName("NonExistent", 10)

	assert.Error(t, err)
}

// ============================================================================
// ServiceFetchTemperatureReadingFromSensor tests
// ============================================================================

func TestSensorService_ServiceFetchTemperatureReadingFromSensor_Success(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/temperature", r.URL.Path)
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := types.Sensor{Id: 1, Name: "TestSensor", URL: server.URL}
	sensorRepo.On("UpdateSensorHealthById", 1, types.SensorGoodHealth, "successful reading").Return(nil)
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{sensor}, nil).Maybe()

	reading, err := service.ServiceFetchTemperatureReadingFromSensor(sensor)

	assert.NoError(t, err)
	assert.Equal(t, 22.5, reading.Temperature)
	assert.Equal(t, "TestSensor", reading.SensorName)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceFetchTemperatureReadingFromSensor_HTTPError(t *testing.T) {
	service, _, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "TestSensor", URL: "http://invalid-url-that-does-not-exist:99999"}

	_, err := service.ServiceFetchTemperatureReadingFromSensor(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error making GET request")
}

func TestSensorService_ServiceFetchTemperatureReadingFromSensor_Non200(t *testing.T) {
	service, _, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sensor := types.Sensor{Id: 1, Name: "TestSensor", URL: server.URL}

	_, err := service.ServiceFetchTemperatureReadingFromSensor(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "received non-200 response")
}

func TestSensorService_ServiceFetchTemperatureReadingFromSensor_InvalidJSON(t *testing.T) {
	service, _, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	sensor := types.Sensor{Id: 1, Name: "TestSensor", URL: server.URL}

	_, err := service.ServiceFetchTemperatureReadingFromSensor(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding JSON")
}

// ============================================================================
// ServiceValidateSensorConfig tests
// ============================================================================

func TestSensorService_ServiceValidateSensorConfig_Valid(t *testing.T) {
	service, sensorRepo, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.RawTempReading{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := types.Sensor{Name: "TestSensor", Type: "Temperature", URL: server.URL}
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors").Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceValidateSensorConfig(sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceValidateSensorConfig_EmptyFields(t *testing.T) {
	service, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "", Type: "Temperature", URL: "http://localhost"}

	err := service.ServiceValidateSensorConfig(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestSensorService_ServiceValidateSensorConfig_FetchFails(t *testing.T) {
	service, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", Type: "Temperature", URL: "http://invalid-url:99999"}

	err := service.ServiceValidateSensorConfig(sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect a reading")
}

// ============================================================================
// AlreadyExistsError tests
// ============================================================================

func TestAlreadyExistsError_Error(t *testing.T) {
	err := NewAlreadyExistsError("sensor already exists")

	assert.Equal(t, "sensor already exists", err.Error())
}
