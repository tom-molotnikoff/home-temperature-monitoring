package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"example/sensorHub/drivers"
	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock repositories
// ============================================================================

type MockMQTTBrokerRepo struct{ mock.Mock }

func (m *MockMQTTBrokerRepo) Add(ctx context.Context, broker types.MQTTBroker) error {
	return m.Called(ctx, broker).Error(0)
}
func (m *MockMQTTBrokerRepo) GetByID(ctx context.Context, id int) (*types.MQTTBroker, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTBroker), args.Error(1)
}
func (m *MockMQTTBrokerRepo) GetByName(ctx context.Context, name string) (*types.MQTTBroker, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTBroker), args.Error(1)
}
func (m *MockMQTTBrokerRepo) GetAll(ctx context.Context) ([]types.MQTTBroker, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.MQTTBroker), args.Error(1)
}
func (m *MockMQTTBrokerRepo) GetEnabled(ctx context.Context) ([]types.MQTTBroker, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.MQTTBroker), args.Error(1)
}
func (m *MockMQTTBrokerRepo) Update(ctx context.Context, broker types.MQTTBroker) error {
	return m.Called(ctx, broker).Error(0)
}
func (m *MockMQTTBrokerRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

type MockMQTTSubRepo struct{ mock.Mock }

func (m *MockMQTTSubRepo) Add(ctx context.Context, sub types.MQTTSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockMQTTSubRepo) GetByID(ctx context.Context, id int) (*types.MQTTSubscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTSubscription), args.Error(1)
}
func (m *MockMQTTSubRepo) GetAll(ctx context.Context) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *MockMQTTSubRepo) GetByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *MockMQTTSubRepo) GetEnabledByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *MockMQTTSubRepo) Update(ctx context.Context, sub types.MQTTSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockMQTTSubRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

// ============================================================================
// Stub PushDriver for subscription validation tests
// ============================================================================

func init() {
	drivers.Register(&stubPushDriver{})
	drivers.Register(&stubNonPushDriver{})
}

type stubPushDriver struct{}

func (d *stubPushDriver) Type() string        { return "mqtt-test-driver" }
func (d *stubPushDriver) DisplayName() string  { return "Test MQTT Driver" }
func (d *stubPushDriver) Description() string  { return "A stub MQTT push driver" }
func (d *stubPushDriver) ConfigFields() []drivers.ConfigFieldSpec { return nil }
func (d *stubPushDriver) SupportedMeasurementTypes() []types.MeasurementType { return nil }
func (d *stubPushDriver) ValidateSensor(_ context.Context, _ types.Sensor) error { return nil }
func (d *stubPushDriver) ParseMessage(_ string, _ []byte) ([]types.Reading, error) { return nil, nil }
func (d *stubPushDriver) IdentifyDevice(_ string, _ []byte) (string, error) { return "", nil }

func setupMQTTService() (*MQTTService, *MockMQTTBrokerRepo, *MockMQTTSubRepo) {
	brokerRepo := new(MockMQTTBrokerRepo)
	subRepo := new(MockMQTTSubRepo)
	svc := NewMQTTService(brokerRepo, subRepo, slog.Default())
	return svc, brokerRepo, subRepo
}

// ============================================================================
// Broker tests
// ============================================================================

func TestMQTTService_AddBroker_Success(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()

	broker := types.MQTTBroker{Name: "test", Type: "external", Host: "mqtt.example.com", Port: 1883, Enabled: true}
	brokerRepo.On("Add", mock.Anything, broker).Return(nil)

	err := svc.AddBroker(context.Background(), broker)
	assert.NoError(t, err)
	brokerRepo.AssertExpectations(t)
}

func TestMQTTService_AddBroker_ValidationFails(t *testing.T) {
	svc, _, _ := setupMQTTService()

	tests := []struct {
		name   string
		broker types.MQTTBroker
		errMsg string
	}{
		{"empty name", types.MQTTBroker{Type: "external", Host: "h", Port: 1883}, "broker name cannot be empty"},
		{"empty host", types.MQTTBroker{Name: "n", Type: "external", Port: 1883}, "broker host cannot be empty"},
		{"invalid port", types.MQTTBroker{Name: "n", Type: "external", Host: "h", Port: 0}, "broker port must be between"},
		{"invalid type", types.MQTTBroker{Name: "n", Type: "invalid", Host: "h", Port: 1883}, "broker type must be"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AddBroker(context.Background(), tt.broker)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestMQTTService_GetAllBrokers(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()

	expected := []types.MQTTBroker{{Id: 1, Name: "b1"}, {Id: 2, Name: "b2"}}
	brokerRepo.On("GetAll", mock.Anything).Return(expected, nil)

	brokers, err := svc.GetAllBrokers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, brokers, 2)
}

func TestMQTTService_UpdateBroker_Success(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()

	broker := types.MQTTBroker{Id: 1, Name: "updated", Type: "external", Host: "h", Port: 1883}
	brokerRepo.On("Update", mock.Anything, broker).Return(nil)

	err := svc.UpdateBroker(context.Background(), broker)
	assert.NoError(t, err)
}

func TestMQTTService_UpdateBroker_InvalidID(t *testing.T) {
	svc, _, _ := setupMQTTService()

	err := svc.UpdateBroker(context.Background(), types.MQTTBroker{Name: "x", Type: "external", Host: "h", Port: 1883})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broker id must be positive")
}

func TestMQTTService_DeleteBroker(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()

	brokerRepo.On("Delete", mock.Anything, 1).Return(nil)

	err := svc.DeleteBroker(context.Background(), 1)
	assert.NoError(t, err)
}

// ============================================================================
// Subscription tests
// ============================================================================

func TestMQTTService_AddSubscription_Success(t *testing.T) {
	svc, brokerRepo, subRepo := setupMQTTService()

	sub := types.MQTTSubscription{BrokerId: 1, TopicPattern: "zigbee2mqtt/+", DriverType: "mqtt-test-driver", Enabled: true}
	brokerRepo.On("GetByID", mock.Anything, 1).Return(&types.MQTTBroker{Id: 1}, nil)
	subRepo.On("Add", mock.Anything, sub).Return(nil)

	err := svc.AddSubscription(context.Background(), sub)
	assert.NoError(t, err)
	subRepo.AssertExpectations(t)
}

func TestMQTTService_AddSubscription_UnknownDriver(t *testing.T) {
	svc, _, _ := setupMQTTService()

	sub := types.MQTTSubscription{BrokerId: 1, TopicPattern: "test/+", DriverType: "nonexistent", Enabled: true}

	err := svc.AddSubscription(context.Background(), sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown driver type")
}

func TestMQTTService_AddSubscription_NotPushDriver(t *testing.T) {
	svc, _, _ := setupMQTTService()

	sub := types.MQTTSubscription{BrokerId: 1, TopicPattern: "test/+", DriverType: "non-push-driver", Enabled: true}

	err := svc.AddSubscription(context.Background(), sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an MQTT push driver")
}

func TestMQTTService_AddSubscription_BrokerNotFound(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()

	sub := types.MQTTSubscription{BrokerId: 99, TopicPattern: "test/+", DriverType: "mqtt-test-driver", Enabled: true}
	brokerRepo.On("GetByID", mock.Anything, 99).Return(nil, errors.New("not found"))

	err := svc.AddSubscription(context.Background(), sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broker not found")
}

func TestMQTTService_AddSubscription_InvalidTopic(t *testing.T) {
	svc, brokerRepo, _ := setupMQTTService()
	brokerRepo.On("GetByID", mock.Anything, 1).Return(&types.MQTTBroker{Id: 1}, nil)

	tests := []struct {
		name    string
		pattern string
		errMsg  string
	}{
		{"spaces", "test topic/+", "must not contain spaces"},
		{"mid hash", "test/#/more", "multi-level wildcard (#) must be the last segment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := types.MQTTSubscription{BrokerId: 1, TopicPattern: tt.pattern, DriverType: "mqtt-test-driver", Enabled: true}
			err := svc.AddSubscription(context.Background(), sub)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestMQTTService_GetAllSubscriptions(t *testing.T) {
	svc, _, subRepo := setupMQTTService()

	expected := []types.MQTTSubscription{{Id: 1}, {Id: 2}}
	subRepo.On("GetAll", mock.Anything).Return(expected, nil)

	subs, err := svc.GetAllSubscriptions(context.Background())
	assert.NoError(t, err)
	assert.Len(t, subs, 2)
}

func TestMQTTService_DeleteSubscription(t *testing.T) {
	svc, _, subRepo := setupMQTTService()

	subRepo.On("Delete", mock.Anything, 1).Return(nil)

	err := svc.DeleteSubscription(context.Background(), 1)
	assert.NoError(t, err)
}

// ============================================================================
// Topic validation tests
// ============================================================================

func TestValidateTopicPattern_Valid(t *testing.T) {
	valid := []string{
		"zigbee2mqtt/+",
		"rtl_433/+/+",
		"tele/+/SENSOR",
		"home/#",
		"+/+/temperature",
	}
	for _, p := range valid {
		assert.NoError(t, validateTopicPattern(p), "expected valid: %s", p)
	}
}

func TestValidateTopicPattern_Invalid(t *testing.T) {
	assert.Error(t, validateTopicPattern("test topic"))
	assert.Error(t, validateTopicPattern("test/#/more"))
}

// ============================================================================
// Stub non-push driver (only implements SensorDriver, not PushDriver)
// ============================================================================

type stubNonPushDriver struct{}

func (d *stubNonPushDriver) Type() string        { return "non-push-driver" }
func (d *stubNonPushDriver) DisplayName() string  { return "Non Push" }
func (d *stubNonPushDriver) Description() string  { return "Not a push driver" }
func (d *stubNonPushDriver) ConfigFields() []drivers.ConfigFieldSpec { return nil }
func (d *stubNonPushDriver) SupportedMeasurementTypes() []types.MeasurementType { return nil }
func (d *stubNonPushDriver) ValidateSensor(_ context.Context, _ types.Sensor) error { return nil }
