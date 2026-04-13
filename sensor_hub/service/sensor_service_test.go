package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example/sensorHub/alerting"
	_ "example/sensorHub/drivers" // trigger init() to register drivers
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

func (m *MockAlertRepository) GetAlertRuleBySensorID(ctx context.Context, sensorID int) (*alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRuleByID(ctx context.Context, ruleID int) (*alerting.AlertRule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRulesBySensorID(ctx context.Context, sensorID int) ([]alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID)
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRuleForReading(ctx context.Context, sensorID int, measurementTypeName string) (*alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID, measurementTypeName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) UpdateLastAlertSent(ctx context.Context, ruleID int) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) RecordAlertSent(ctx context.Context, ruleID, sensorID, measurementTypeId int, reason string, numericValue float64, statusValue string) error {
	args := m.Called(ctx, ruleID, sensorID, measurementTypeId, reason, numericValue, statusValue)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAllAlertRules(ctx context.Context) ([]alerting.AlertRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRuleBySensorName(ctx context.Context, sensorName string) (*alerting.AlertRule, error) {
	args := m.Called(ctx, sensorName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

func (m *MockAlertRepository) CreateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertRepository) UpdateAlertRule(ctx context.Context, rule *alerting.AlertRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAlertRepository) DeleteAlertRule(ctx context.Context, ruleID int) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockAlertRepository) GetAlertHistory(ctx context.Context, sensorID int, limit int) ([]types.AlertHistoryEntry, error) {
	args := m.Called(ctx, sensorID, limit)
	return args.Get(0).([]types.AlertHistoryEntry), args.Error(1)
}

func (m *MockAlertRepository) GetAlertRule(ctx context.Context, sensorID, measurementTypeId int) (*alerting.AlertRule, error) {
	args := m.Called(ctx, sensorID, measurementTypeId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alerting.AlertRule), args.Error(1)
}

// ============================================================================
// Mock MeasurementTypeRepository for SensorService
// ============================================================================

type MockMeasurementTypeRepository struct {
	mock.Mock
}

func (m *MockMeasurementTypeRepository) GetAll(ctx context.Context) ([]types.MeasurementType, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.MeasurementType), args.Error(1)
}

func (m *MockMeasurementTypeRepository) GetByName(ctx context.Context, name string) (*types.MeasurementType, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MeasurementType), args.Error(1)
}

func (m *MockMeasurementTypeRepository) GetBySensorId(ctx context.Context, sensorId int) ([]types.SensorMeasurementType, error) {
	args := m.Called(ctx, sensorId)
	return args.Get(0).([]types.SensorMeasurementType), args.Error(1)
}

func (m *MockMeasurementTypeRepository) EnsureExists(ctx context.Context, mt types.MeasurementType) error {
	args := m.Called(ctx, mt)
	return args.Error(0)
}

func (m *MockMeasurementTypeRepository) AssignToSensor(ctx context.Context, sensorId, measurementTypeId int, unit string) error {
	args := m.Called(ctx, sensorId, measurementTypeId, unit)
	return args.Error(0)
}

func (m *MockMeasurementTypeRepository) RemoveFromSensor(ctx context.Context, sensorId, measurementTypeId int) error {
	args := m.Called(ctx, sensorId, measurementTypeId)
	return args.Error(0)
}

func (m *MockMeasurementTypeRepository) GetMeasurementTypesWithReadings(ctx context.Context, sensorId int) ([]types.MeasurementType, error) {
	args := m.Called(ctx, sensorId)
	return args.Get(0).([]types.MeasurementType), args.Error(1)
}

func (m *MockMeasurementTypeRepository) GetAllWithReadings(ctx context.Context) ([]types.MeasurementType, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.MeasurementType), args.Error(1)
}

// ============================================================================
// Test helpers
// ============================================================================

func setupSensorService() (*SensorService, *MockSensorRepository, *MockReadingsRepository, *MockMeasurementTypeRepository, *MockAlertRepository) {
	sensorRepo := new(MockSensorRepository)
	readingsRepo := new(MockReadingsRepository)
	mtRepo := new(MockMeasurementTypeRepository)
	alertRepo := new(MockAlertRepository)

	service := NewSensorService(sensorRepo, readingsRepo, mtRepo, alertRepo, nil, slog.Default())
	return service, sensorRepo, readingsRepo, mtRepo, alertRepo
}

// ============================================================================
// ServiceAddSensor tests
// ============================================================================

func TestSensorService_ServiceAddSensor_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}, Enabled: true}

	// Create a mock HTTP server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.Config = map[string]string{"url": server.URL}

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(false, nil)
	sensorRepo.On("AddSensor", mock.Anything,  mock.Anything).Return(nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceAddSensor(context.Background(), sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond) // Allow async goroutine to complete
}

func TestSensorService_ServiceAddSensor_AlreadyExists(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "ExistingSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}

	// Create mock server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.Config = map[string]string{"url": server.URL}

	sensorRepo.On("SensorExists", mock.Anything, "ExistingSensor").Return(true, nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceAddSensor(context.Background(), sensor)

	assert.Error(t, err)
	var alreadyExistsErr *AlreadyExistsError
	assert.True(t, errors.As(err, &alreadyExistsErr))
}

func TestSensorService_ServiceAddSensor_ValidationError_EmptyName(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}

	err := service.ServiceAddSensor(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor validation failed")
}

func TestSensorService_ServiceAddSensor_SensorExistsError(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}

	// Create mock server for validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.Config = map[string]string{"url": server.URL}

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(false, errors.New("database error"))
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceAddSensor(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error checking if sensor exists")
}

// ============================================================================
// ServiceUpdateSensorById tests
// ============================================================================

func TestSensorService_ServiceUpdateSensorById_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "UpdatedSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.Config = map[string]string{"url": server.URL}

	sensorRepo.On("UpdateSensorById", mock.Anything,  mock.Anything).Return(nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceUpdateSensorById(context.Background(), sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceUpdateSensorById_ValidationError(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost:8080"}}

	err := service.ServiceUpdateSensorById(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor validation failed")
}

// ============================================================================
// ServiceDeleteSensorByName tests
// ============================================================================

func TestSensorService_ServiceDeleteSensorByName_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(true, nil)
	sensorRepo.On("DeleteSensorByName", mock.Anything, "TestSensor").Return(nil)
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceDeleteSensorByName(context.Background(), "TestSensor")

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceDeleteSensorByName_NotExists(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "NonExistent").Return(false, nil)

	err := service.ServiceDeleteSensorByName(context.Background(), "NonExistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestSensorService_ServiceDeleteSensorByName_Error(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(true, nil)
	sensorRepo.On("DeleteSensorByName", mock.Anything, "TestSensor").Return(errors.New("database error"))

	err := service.ServiceDeleteSensorByName(context.Background(), "TestSensor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting sensor")
}

// ============================================================================
// ServiceGetSensorByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorByName_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensor := &types.Sensor{Id: 1, Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature"}
	sensorRepo.On("GetSensorByName", mock.Anything, "TestSensor").Return(sensor, nil)

	result, err := service.ServiceGetSensorByName(context.Background(), "TestSensor")

	assert.NoError(t, err)
	assert.Equal(t, "TestSensor", result.Name)
}

func TestSensorService_ServiceGetSensorByName_EmptyName(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	result, err := service.ServiceGetSensorByName(context.Background(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sensor name cannot be empty")
	assert.Nil(t, result)
}

func TestSensorService_ServiceGetSensorByName_NotFound(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetSensorByName", mock.Anything, "NonExistent").Return(nil, nil)

	result, err := service.ServiceGetSensorByName(context.Background(), "NonExistent")

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// ServiceGetAllSensors tests
// ============================================================================

func TestSensorService_ServiceGetAllSensors_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "Sensor1"},
		{Id: 2, Name: "Sensor2"},
	}
	sensorRepo.On("GetAllSensors", mock.Anything).Return(sensors, nil)

	result, err := service.ServiceGetAllSensors(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestSensorService_ServiceGetAllSensors_Empty(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, nil)

	result, err := service.ServiceGetAllSensors(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestSensorService_ServiceGetAllSensors_Error(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, errors.New("database error"))

	_, err := service.ServiceGetAllSensors(context.Background())

	assert.Error(t, err)
}

// ============================================================================
// ServiceGetSensorsByDriver tests
// ============================================================================

func TestSensorService_ServiceGetSensorsByDriver_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "TempSensor1", SensorDriver: "sensor-hub-http-temperature"},
		{Id: 2, Name: "TempSensor2", SensorDriver: "sensor-hub-http-temperature"},
	}
	sensorRepo.On("GetSensorsByDriver", mock.Anything, "sensor-hub-http-temperature").Return(sensors, nil)

	result, err := service.ServiceGetSensorsByDriver(context.Background(), "sensor-hub-http-temperature")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// ============================================================================
// ServiceGetSensorIdByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorIdByName_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", mock.Anything, "TestSensor").Return(1, nil)

	result, err := service.ServiceGetSensorIdByName(context.Background(), "TestSensor")

	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestSensorService_ServiceGetSensorIdByName_Error(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", mock.Anything, "NonExistent").Return(0, errors.New("not found"))

	_, err := service.ServiceGetSensorIdByName(context.Background(), "NonExistent")

	assert.Error(t, err)
}

// ============================================================================
// ServiceSensorExists tests
// ============================================================================

func TestSensorService_ServiceSensorExists_True(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "ExistingSensor").Return(true, nil)

	result, err := service.ServiceSensorExists(context.Background(), "ExistingSensor")

	assert.NoError(t, err)
	assert.True(t, result)
}

func TestSensorService_ServiceSensorExists_False(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "NonExistent").Return(false, nil)

	result, err := service.ServiceSensorExists(context.Background(), "NonExistent")

	assert.NoError(t, err)
	assert.False(t, result)
}

// ============================================================================
// ServiceSetEnabledSensorByName tests
// ============================================================================

func TestSensorService_ServiceSetEnabledSensorByName_Enable(t *testing.T) {
	service, sensorRepo, readingsRepo, _, alertRepo := setupSensorService()

	sensor := &types.Sensor{Id: 1, Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature", Enabled: true}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()
	sensor.Config = map[string]string{"url": server.URL}

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(true, nil)
	sensorRepo.On("SetEnabledSensorByName", mock.Anything, "TestSensor", true).Return(nil)
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{*sensor}, nil).Maybe()
	sensorRepo.On("GetSensorByName", mock.Anything, "TestSensor").Return(sensor, nil).Maybe()
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	readingsRepo.On("Add", mock.Anything,  mock.Anything).Return(nil).Maybe()
	// The async collection triggers alert processing
	alertRepo.On("GetAlertRuleForReading", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	err := service.ServiceSetEnabledSensorByName(context.Background(), "TestSensor", true)

	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
}

func TestSensorService_ServiceSetEnabledSensorByName_Disable(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "TestSensor").Return(true, nil)
	sensorRepo.On("SetEnabledSensorByName", mock.Anything, "TestSensor", false).Return(nil)
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{}, nil).Maybe()

	err := service.ServiceSetEnabledSensorByName(context.Background(), "TestSensor", false)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceSetEnabledSensorByName_NotExists(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("SensorExists", mock.Anything, "NonExistent").Return(false, nil)

	err := service.ServiceSetEnabledSensorByName(context.Background(), "NonExistent", true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// ============================================================================
// ServiceGetTotalReadingsForEachSensor tests
// ============================================================================

func TestSensorService_ServiceGetTotalReadingsForEachSensor_Success(t *testing.T) {
	service, sensorRepo, readingsRepo, _, _ := setupSensorService()

	sensors := []types.Sensor{
		{Id: 1, Name: "Sensor1", SensorDriver: "sensor-hub-http-temperature"},
		{Id: 2, Name: "Sensor2", SensorDriver: "sensor-hub-http-temperature"},
	}
	sensorRepo.On("GetSensorsByStatus", mock.Anything, "active").Return(sensors, nil)
	readingsRepo.On("GetTotalReadingsBySensorId", mock.Anything, 1).Return(100, nil)
	readingsRepo.On("GetTotalReadingsBySensorId", mock.Anything, 2).Return(50, nil)

	result, err := service.ServiceGetTotalReadingsForEachSensor(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 100, result["Sensor1"])
	assert.Equal(t, 50, result["Sensor2"])
}

func TestSensorService_ServiceGetTotalReadingsForEachSensor_Error(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetSensorsByStatus", mock.Anything, "active").Return([]types.Sensor{}, errors.New("database error"))

	_, err := service.ServiceGetTotalReadingsForEachSensor(context.Background())

	assert.Error(t, err)
}

// ============================================================================
// ServiceGetSensorHealthHistoryByName tests
// ============================================================================

func TestSensorService_ServiceGetSensorHealthHistoryByName_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	history := []types.SensorHealthHistory{
		{SensorId: "1", HealthStatus: types.SensorGoodHealth},
	}
	sensorRepo.On("GetSensorIdByName", mock.Anything, "TestSensor").Return(1, nil)
	sensorRepo.On("GetSensorHealthHistoryById", mock.Anything, 1, 10).Return(history, nil)

	result, err := service.ServiceGetSensorHealthHistoryByName(context.Background(), "TestSensor", 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSensorService_ServiceGetSensorHealthHistoryByName_SensorNotFound(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	sensorRepo.On("GetSensorIdByName", mock.Anything, "NonExistent").Return(0, errors.New("not found"))

	_, err := service.ServiceGetSensorHealthHistoryByName(context.Background(), "NonExistent", 10)

	assert.Error(t, err)
}

// ============================================================================
// ServiceValidateSensorConfig tests
// ============================================================================

func TestSensorService_ServiceValidateSensorConfig_Valid(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := types.Sensor{Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": server.URL}}
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceValidateSensorConfig(context.Background(), sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceValidateSensorConfig_EmptyFields(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost"}}

	err := service.ServiceValidateSensorConfig(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestSensorService_ServiceValidateSensorConfig_FetchFails(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	sensor := types.Sensor{Name: "TestSensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://invalid-url:99999"}}

	err := service.ServiceValidateSensorConfig(context.Background(), sensor)

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

// ============================================================================
// ServiceCollectFromSensorByName contract tests
// ============================================================================

func TestSensorService_ServiceCollectFromSensorByName_Success(t *testing.T) {
	service, sensorRepo, readingsRepo, _, alertRepo := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := &types.Sensor{Id: 1, Name: "test-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": server.URL}, Enabled: true}
	sensorRepo.On("GetSensorByName", mock.Anything, "test-sensor").Return(sensor, nil)
	readingsRepo.On("Add", mock.Anything, mock.Anything).Return(nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, 1, types.SensorGoodHealth, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{*sensor}, nil).Maybe()
	alertRepo.On("GetAlertRuleForReading", mock.Anything, 1, mock.Anything).Return(nil, nil).Maybe()

	err := service.ServiceCollectFromSensorByName(context.Background(), "test-sensor")

	assert.NoError(t, err)
	readingsRepo.AssertCalled(t, "Add", mock.Anything, mock.MatchedBy(func(readings []types.Reading) bool {
		return len(readings) == 1 && readings[0].NumericValue != nil && *readings[0].NumericValue == 22.5 && readings[0].SensorName == "test-sensor"
	}))
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceCollectFromSensorByName_SensorNotFound(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()
	sensorRepo.On("GetSensorByName", mock.Anything, "missing").Return(nil, nil)

	err := service.ServiceCollectFromSensorByName(context.Background(), "missing")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSensorService_ServiceCollectFromSensorByName_DisabledSensor(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()
	sensor := &types.Sensor{Id: 1, Name: "disabled-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": "http://localhost"}, Enabled: false}
	sensorRepo.On("GetSensorByName", mock.Anything, "disabled-sensor").Return(sensor, nil)

	err := service.ServiceCollectFromSensorByName(context.Background(), "disabled-sensor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestSensorService_ServiceCollectFromSensorByName_UnsupportedType(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()
	sensor := &types.Sensor{Id: 1, Name: "unknown-sensor", SensorDriver: "humidity", Config: map[string]string{"url": "http://localhost"}, Enabled: true}
	sensorRepo.On("GetSensorByName", mock.Anything, "unknown-sensor").Return(sensor, nil)

	err := service.ServiceCollectFromSensorByName(context.Background(), "unknown-sensor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported sensor driver")
}

func TestSensorService_ServiceCollectFromSensorByName_FetchError_SetsHealthBad(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sensor := &types.Sensor{Id: 1, Name: "failing-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": server.URL}, Enabled: true}
	sensorRepo.On("GetSensorByName", mock.Anything, "failing-sensor").Return(sensor, nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, 1, types.SensorBadHealth, mock.Anything).Return(nil)
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{*sensor}, nil).Maybe()

	err := service.ServiceCollectFromSensorByName(context.Background(), "failing-sensor")

	assert.Error(t, err)
	sensorRepo.AssertCalled(t, "UpdateSensorHealthById", mock.Anything, 1, types.SensorBadHealth, mock.Anything)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceCollectFromSensorByName_StoreError_SetsHealthBad(t *testing.T) {
	service, sensorRepo, readingsRepo, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 22.5, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := &types.Sensor{Id: 1, Name: "store-fail-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": server.URL}, Enabled: true}
	sensorRepo.On("GetSensorByName", mock.Anything, "store-fail-sensor").Return(sensor, nil)
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, 1, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{*sensor}, nil).Maybe()
	readingsRepo.On("Add", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := service.ServiceCollectFromSensorByName(context.Background(), "store-fail-sensor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error storing")
	time.Sleep(50 * time.Millisecond)
}

// ============================================================================
// ServiceCollectReadingToValidateSensor contract tests
// ============================================================================

func TestSensorService_ServiceCollectReadingToValidateSensor_Success(t *testing.T) {
	service, sensorRepo, _, _, _ := setupSensorService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"temperature": 20.0, "time": "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	sensor := types.Sensor{Id: 1, Name: "validate-sensor", SensorDriver: "sensor-hub-http-temperature", Config: map[string]string{"url": server.URL}, Enabled: true}
	sensorRepo.On("UpdateSensorHealthById", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	sensorRepo.On("GetAllSensors", mock.Anything).Return([]types.Sensor{sensor}, nil).Maybe()

	err := service.ServiceCollectReadingToValidateSensor(context.Background(), sensor)

	assert.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestSensorService_ServiceCollectReadingToValidateSensor_UnsupportedType(t *testing.T) {
	service, _, _, _, _ := setupSensorService()

	sensor := types.Sensor{Id: 1, Name: "bad-type", SensorDriver: "humidity", Config: map[string]string{"url": "http://localhost"}, Enabled: true}

	err := service.ServiceCollectReadingToValidateSensor(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported sensor driver")
}
