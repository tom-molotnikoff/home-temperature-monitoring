package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"example/sensorHub/drivers"
	gen "example/sensorHub/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mocks
// ============================================================================

type MockSensorService struct {
	mock.Mock
}

func (m *MockSensorService) ServiceGetSensorByName(ctx context.Context, name string) (*gen.Sensor, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.Sensor), args.Error(1)
}

func (m *MockSensorService) ServiceSensorExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockSensorService) ServiceAddSensor(ctx context.Context, sensor gen.Sensor) error {
	return m.Called(ctx, sensor).Error(0)
}

func (m *MockSensorService) ServiceUpdateSensorHealthById(ctx context.Context, sensorId int, healthStatus gen.SensorHealthStatus, healthReason string) {
	m.Called(ctx, sensorId, healthStatus, healthReason)
}

func (m *MockSensorService) ServiceUpdateSensorById(ctx context.Context, sensor gen.Sensor) error {
	return m.Called(ctx, sensor).Error(0)
}
func (m *MockSensorService) ServiceDeleteSensorByName(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}
func (m *MockSensorService) ServiceGetAllSensors(ctx context.Context) ([]gen.Sensor, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.Sensor), args.Error(1)
}
func (m *MockSensorService) ServiceGetSensorsByDriver(ctx context.Context, sensorDriver string) ([]gen.Sensor, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.Sensor), args.Error(1)
}
func (m *MockSensorService) ServiceGetSensorIdByName(ctx context.Context, name string) (int, error) {
	args := m.Called(ctx, name)
	return args.Int(0), args.Error(1)
}
func (m *MockSensorService) ServiceCollectAndStoreAllSensorReadings(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockSensorService) ServiceCollectFromSensorByName(ctx context.Context, sensorName string) error {
	return m.Called(ctx, sensorName).Error(0)
}
func (m *MockSensorService) ServiceCollectReadingToValidateSensor(ctx context.Context, sensor gen.Sensor) error {
	return m.Called(ctx, sensor).Error(0)
}
func (m *MockSensorService) ServiceStartPeriodicSensorCollection(ctx context.Context) {
	m.Called(ctx)
}
func (m *MockSensorService) ServiceDiscoverSensors(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockSensorService) ServiceValidateSensorConfig(ctx context.Context, sensor gen.Sensor) error {
	return m.Called(ctx, sensor).Error(0)
}
func (m *MockSensorService) ServiceSetEnabledSensorByName(ctx context.Context, name string, enabled bool) error {
	return m.Called(ctx, name, enabled).Error(0)
}
func (m *MockSensorService) ServiceGetSensorHealthHistoryByName(ctx context.Context, name string, limit int) ([]gen.SensorHealthHistory, error) {
	args := m.Called(ctx, name, limit)
	return args.Get(0).([]gen.SensorHealthHistory), args.Error(1)
}
func (m *MockSensorService) ServiceGetTotalReadingsForEachSensor(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int), args.Error(1)
}
func (m *MockSensorService) ServiceGetSensorsByStatus(ctx context.Context, status string) ([]gen.Sensor, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]gen.Sensor), args.Error(1)
}
func (m *MockSensorService) ServiceApproveSensor(ctx context.Context, sensorId int) error {
	return m.Called(ctx, sensorId).Error(0)
}
func (m *MockSensorService) ServiceDismissSensor(ctx context.Context, sensorId int) error {
	return m.Called(ctx, sensorId).Error(0)
}
func (m *MockSensorService) ServiceProcessPushReadings(ctx context.Context, sensor gen.Sensor, readings []gen.Reading) error {
	return m.Called(ctx, sensor, readings).Error(0)
}
func (m *MockSensorService) ServiceGetMeasurementTypesForSensor(ctx context.Context, sensorId int) ([]gen.MeasurementType, error) {
	args := m.Called(ctx, sensorId)
	return args.Get(0).([]gen.MeasurementType), args.Error(1)
}
func (m *MockSensorService) ServiceGetAllMeasurementTypes(ctx context.Context) ([]gen.MeasurementType, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.MeasurementType), args.Error(1)
}
func (m *MockSensorService) ServiceGetSensorByExternalId(ctx context.Context, externalId string) (*gen.Sensor, error) {
	args := m.Called(ctx, externalId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.Sensor), args.Error(1)
}
func (m *MockSensorService) ServiceGetSensorById(ctx context.Context, id int) (*gen.Sensor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.Sensor), args.Error(1)
}
func (m *MockSensorService) ServiceSensorExistsByExternalId(ctx context.Context, externalId string) (bool, error) {
	args := m.Called(ctx, externalId)
	return args.Bool(0), args.Error(1)
}
func (m *MockSensorService) ServiceGetAllMeasurementTypesWithReadings(ctx context.Context) ([]gen.MeasurementType, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.MeasurementType), args.Error(1)
}

type MockSubRepo struct {
	mock.Mock
}

func (m *MockSubRepo) Add(ctx context.Context, sub gen.MQTTSubscription) (int, error) {
	args := m.Called(ctx, sub)
	return args.Int(0), args.Error(1)
}
func (m *MockSubRepo) GetByID(ctx context.Context, id int) (*gen.MQTTSubscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.MQTTSubscription), args.Error(1)
}
func (m *MockSubRepo) GetAll(ctx context.Context) ([]gen.MQTTSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.MQTTSubscription), args.Error(1)
}
func (m *MockSubRepo) GetByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	return args.Get(0).([]gen.MQTTSubscription), args.Error(1)
}
func (m *MockSubRepo) GetEnabledByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	return args.Get(0).([]gen.MQTTSubscription), args.Error(1)
}
func (m *MockSubRepo) Update(ctx context.Context, sub gen.MQTTSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockSubRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

type MockBrokerRepo struct {
	mock.Mock
}

func (m *MockBrokerRepo) Add(ctx context.Context, broker gen.MQTTBroker) (int, error) {
	args := m.Called(ctx, broker)
	return args.Int(0), args.Error(1)
}
func (m *MockBrokerRepo) GetByID(ctx context.Context, id int) (*gen.MQTTBroker, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.MQTTBroker), args.Error(1)
}
func (m *MockBrokerRepo) GetByName(ctx context.Context, name string) (*gen.MQTTBroker, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.MQTTBroker), args.Error(1)
}
func (m *MockBrokerRepo) GetAll(ctx context.Context) ([]gen.MQTTBroker, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.MQTTBroker), args.Error(1)
}
func (m *MockBrokerRepo) GetEnabled(ctx context.Context) ([]gen.MQTTBroker, error) {
	args := m.Called(ctx)
	return args.Get(0).([]gen.MQTTBroker), args.Error(1)
}
func (m *MockBrokerRepo) Update(ctx context.Context, broker gen.MQTTBroker) error {
	return m.Called(ctx, broker).Error(0)
}
func (m *MockBrokerRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

// ============================================================================
// Stub PushDriver
// ============================================================================

type stubPushDriver struct{}

func (s *stubPushDriver) Type() string                                                  { return "test-push-driver" }
func (s *stubPushDriver) DisplayName() string                                           { return "Test Push" }
func (s *stubPushDriver) Description() string                                           { return "test" }
func (s *stubPushDriver) ConfigFields() []drivers.ConfigFieldSpec                       { return nil }
func (s *stubPushDriver) SupportedMeasurementTypes() []gen.MeasurementType            { return nil }
func (s *stubPushDriver) ValidateSensor(ctx context.Context, sensor gen.Sensor) error { return nil }

func (s *stubPushDriver) ParseMessage(topic string, payload []byte) ([]gen.Reading, error) {
	val := 22.5
	return []gen.Reading{
		{NumericValue: &val, MeasurementType: "temperature", Unit: "°C"},
	}, nil
}

func (s *stubPushDriver) IdentifyDevice(topic string, payload []byte) (string, error) {
	return "mqtt-device-1", nil
}

func init() {
	drivers.Register(&stubPushDriver{})
}

// ============================================================================
// Tests
// ============================================================================

func TestConnectionManager_HandleMessage_KnownSensor(t *testing.T) {
	mockSensor := &MockSensorService{}
	mockSub := &MockSubRepo{}
	mockBroker := &MockBrokerRepo{}

	cm := NewConnectionManager(mockSensor, mockSub, mockBroker, slog.Default())

	sensor := &gen.Sensor{Id: 1, Name: "mqtt-device-1", SensorDriver: "test-push-driver", Status: gen.SensorStatusActive, Enabled: true}
	mockSensor.On("ServiceGetSensorByExternalId", mock.Anything, "mqtt-device-1").Return(sensor, nil)
	mockSensor.On("ServiceProcessPushReadings", mock.Anything, *sensor, mock.AnythingOfType("[]gen.Reading")).Return(nil)

	cm.handleMessage(context.Background(), 1, "test-push-driver", "test/topic", []byte(`{}`))

	mockSensor.AssertExpectations(t)
}

func TestConnectionManager_HandleMessage_AutoDiscovery(t *testing.T) {
	mockSensor := &MockSensorService{}
	mockSub := &MockSubRepo{}
	mockBroker := &MockBrokerRepo{}

	cm := NewConnectionManager(mockSensor, mockSub, mockBroker, slog.Default())

	mockSensor.On("ServiceGetSensorByExternalId", mock.Anything, "mqtt-device-1").Return(nil, fmt.Errorf("not found"))
	mockSensor.On("ServiceGetSensorByName", mock.Anything, "mqtt-device-1").Return(nil, fmt.Errorf("not found"))
	mockSensor.On("ServiceSensorExistsByExternalId", mock.Anything, "mqtt-device-1").Return(false, nil)
	mockSensor.On("ServiceSensorExists", mock.Anything, "mqtt-device-1").Return(false, nil)
	mockSensor.On("ServiceAddSensor", mock.Anything, mock.MatchedBy(func(s gen.Sensor) bool {
		return s.Name == "mqtt-device-1" && s.Status == gen.SensorStatusPending && s.ExternalId != nil && *s.ExternalId == "mqtt-device-1"
	})).Return(nil)

	cm.handleMessage(context.Background(), 1, "test-push-driver", "test/topic", []byte(`{}`))

	mockSensor.AssertExpectations(t)
}

func TestConnectionManager_HandleMessage_InactiveSensor(t *testing.T) {
	mockSensor := &MockSensorService{}
	mockSub := &MockSubRepo{}
	mockBroker := &MockBrokerRepo{}

	cm := NewConnectionManager(mockSensor, mockSub, mockBroker, slog.Default())

	sensor := &gen.Sensor{Id: 2, Name: "mqtt-device-1", Status: gen.SensorStatusPending, Enabled: true}
	mockSensor.On("ServiceGetSensorByExternalId", mock.Anything, "mqtt-device-1").Return(sensor, nil)

	cm.handleMessage(context.Background(), 1, "test-push-driver", "test/topic", []byte(`{}`))

	// Should NOT call ServiceProcessPushReadings since sensor is pending
	mockSensor.AssertNotCalled(t, "ServiceProcessPushReadings")
}

func TestConnectionManager_HandleMessage_UnknownDriver(t *testing.T) {
	mockSensor := &MockSensorService{}
	mockSub := &MockSubRepo{}
	mockBroker := &MockBrokerRepo{}

	cm := NewConnectionManager(mockSensor, mockSub, mockBroker, slog.Default())

	cm.handleMessage(context.Background(), 1, "nonexistent-driver", "test/topic", []byte(`{}`))

	// No service calls should be made
	mockSensor.AssertNotCalled(t, "ServiceGetSensorByName")
}

func TestConnectionManager_HandleMessage_DisabledSensor(t *testing.T) {
	mockSensor := &MockSensorService{}
	mockSub := &MockSubRepo{}
	mockBroker := &MockBrokerRepo{}

	cm := NewConnectionManager(mockSensor, mockSub, mockBroker, slog.Default())

	sensor := &gen.Sensor{Id: 3, Name: "mqtt-device-1", Status: gen.SensorStatusActive, Enabled: false}
	mockSensor.On("ServiceGetSensorByExternalId", mock.Anything, "mqtt-device-1").Return(sensor, nil)

	cm.handleMessage(context.Background(), 1, "test-push-driver", "test/topic", []byte(`{}`))

	// Should NOT process readings for disabled sensors
	mockSensor.AssertNotCalled(t, "ServiceProcessPushReadings")
}

func TestConnectionManager_IsConnected_NoConnection(t *testing.T) {
	cm := NewConnectionManager(nil, nil, nil, slog.Default())
	assert.False(t, cm.IsConnected(999))
}

func TestConnectionManager_ConnectedBrokerIDs_Empty(t *testing.T) {
	cm := NewConnectionManager(nil, nil, nil, slog.Default())
	assert.Empty(t, cm.ConnectedBrokerIDs())
}
