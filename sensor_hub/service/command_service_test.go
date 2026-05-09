package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gen "example/sensorHub/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockCommandSensorRepository struct{ mock.Mock }

func (m *mockCommandSensorRepository) GetSensorById(ctx context.Context, id int) (*gen.Sensor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.Sensor), args.Error(1)
}

type mockCommandSubscriptionRepository struct{ mock.Mock }

func (m *mockCommandSubscriptionRepository) ListEnabledByDriverType(ctx context.Context, driverType string) ([]gen.MQTTSubscription, error) {
	args := m.Called(ctx, driverType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]gen.MQTTSubscription), args.Error(1)
}

type mockCommandHistoryRepository struct{ mock.Mock }

func (m *mockCommandHistoryRepository) HasPendingCommand(ctx context.Context, sensorID int, property string) (bool, error) {
	args := m.Called(ctx, sensorID, property)
	return args.Bool(0), args.Error(1)
}

func (m *mockCommandHistoryRepository) AddSentCommand(ctx context.Context, sensorID int, userID *int, property string, value string, mqttTopic string, mqttPayload string, timeoutSeconds int, sentAt time.Time) (int, error) {
	args := m.Called(ctx, sensorID, userID, property, value, mqttTopic, mqttPayload, timeoutSeconds, sentAt)
	return args.Int(0), args.Error(1)
}

type mockCommandPublisher struct{ mock.Mock }

func (m *mockCommandPublisher) Publish(brokerID int, topic string, payload []byte, qos byte) error {
	return m.Called(brokerID, topic, payload, qos).Error(0)
}

func TestCommandService_Send_PublishesAndPersistsSentCommand(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensor := &gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
		Metadata: &map[string]interface{}{
			"exposes": []interface{}{
				map[string]interface{}{
					"type":      "binary",
					"property":  "state",
					"access":    float64(7),
					"value_on":  "ON",
					"value_off": "OFF",
				},
			},
		},
	}
	userID := 99
	actor := &gen.User{Id: userID, Permissions: []string{"control_sensors"}}
	sub := gen.MQTTSubscription{BrokerId: 12, DriverType: "mqtt-zigbee2mqtt", Enabled: true}

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(sensor, nil)
	subRepo.On("ListEnabledByDriverType", mock.Anything, "mqtt-zigbee2mqtt").Return([]gen.MQTTSubscription{sub}, nil)
	historyRepo.On("HasPendingCommand", mock.Anything, 7, "state").Return(false, nil)
	publisher.On("Publish", 12, "zigbee2mqtt/office-plug/set", []byte(`{"state":"ON"}`), byte(1)).Return(nil)
	historyRepo.On("AddSentCommand",
		mock.Anything,
		7,
		&userID,
		"state",
		"ON",
		"zigbee2mqtt/office-plug/set",
		`{"state":"ON"}`,
		10,
		mock.AnythingOfType("time.Time"),
	).Return(42, nil)

	result, err := svc.Send(context.Background(), 7, actor, "state", "ON")

	require.NoError(t, err)
	assert.Equal(t, SentCommandResult{
		ID:       42,
		Status:   "sent",
		Property: "state",
		Value:    "ON",
	}, result)
}

func TestCommandService_Send_InsufficientPermission(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
	}, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"view_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusForbidden, commandErr.StatusCode)
}

func TestCommandService_Send_NotControllable(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "living-room",
		SensorDriver: "sensor-hub-http-temperature",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
	}, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusBadRequest, commandErr.StatusCode)
}

func TestCommandService_Send_SensorPending(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusPending,
		Enabled:      true,
	}, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusConflict, commandErr.StatusCode)
}

func TestCommandService_Send_SensorDisabled(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      false,
	}, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusConflict, commandErr.StatusCode)
}

func TestCommandService_Send_InvalidValue(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
		Metadata: &map[string]interface{}{
			"exposes": []interface{}{
				map[string]interface{}{
					"type":      "binary",
					"property":  "state",
					"access":    float64(7),
					"value_on":  "ON",
					"value_off": "OFF",
				},
			},
		},
	}, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "INVALID")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusBadRequest, commandErr.StatusCode)
}

func TestCommandService_Send_DuplicatePendingCommand(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
		Metadata: &map[string]interface{}{
			"exposes": []interface{}{
				map[string]interface{}{
					"type":      "binary",
					"property":  "state",
					"access":    float64(7),
					"value_on":  "ON",
					"value_off": "OFF",
				},
			},
		},
	}, nil)
	historyRepo.On("HasPendingCommand", mock.Anything, 7, "state").Return(true, nil)

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusTooManyRequests, commandErr.StatusCode)
}

func TestCommandService_Send_BrokerDisconnected(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
		Metadata: &map[string]interface{}{
			"exposes": []interface{}{
				map[string]interface{}{
					"type":      "binary",
					"property":  "state",
					"access":    float64(7),
					"value_on":  "ON",
					"value_off": "OFF",
				},
			},
		},
	}, nil)
	subRepo.On("ListEnabledByDriverType", mock.Anything, "mqtt-zigbee2mqtt").Return([]gen.MQTTSubscription{{
		BrokerId:   12,
		DriverType: "mqtt-zigbee2mqtt",
		Enabled:    true,
	}}, nil)
	historyRepo.On("HasPendingCommand", mock.Anything, 7, "state").Return(false, nil)
	publisher.On("Publish", 12, "zigbee2mqtt/office-plug/set", []byte(`{"state":"ON"}`), byte(1)).Return(errors.New("broker 12 is not connected"))

	_, err := svc.Send(context.Background(), 7, &gen.User{Id: 5, Permissions: []string{"control_sensors"}}, "state", "ON")

	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	assert.Equal(t, http.StatusServiceUnavailable, commandErr.StatusCode)
}

func TestCommandService_Send_SkipsDisconnectedSubscriptionAndPublishesToNext(t *testing.T) {
	sensorRepo := &mockCommandSensorRepository{}
	subRepo := &mockCommandSubscriptionRepository{}
	historyRepo := &mockCommandHistoryRepository{}
	publisher := &mockCommandPublisher{}
	svc := NewCommandService(sensorRepo, subRepo, historyRepo, publisher, nil)

	sensorRepo.On("GetSensorById", mock.Anything, 7).Return(&gen.Sensor{
		Id:           7,
		Name:         "office-plug",
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Enabled:      true,
		Metadata: &map[string]interface{}{
			"exposes": []interface{}{
				map[string]interface{}{
					"type":      "binary",
					"property":  "state",
					"access":    float64(7),
					"value_on":  "ON",
					"value_off": "OFF",
				},
			},
		},
	}, nil)
	subRepo.On("ListEnabledByDriverType", mock.Anything, "mqtt-zigbee2mqtt").Return([]gen.MQTTSubscription{
		{BrokerId: 12, DriverType: "mqtt-zigbee2mqtt", Enabled: true},
		{BrokerId: 18, DriverType: "mqtt-zigbee2mqtt", Enabled: true},
	}, nil)
	historyRepo.On("HasPendingCommand", mock.Anything, 7, "state").Return(false, nil)
	publisher.On("Publish", 12, "zigbee2mqtt/office-plug/set", []byte(`{"state":"ON"}`), byte(1)).Return(errors.New("broker 12 is not connected"))
	publisher.On("Publish", 18, "zigbee2mqtt/office-plug/set", []byte(`{"state":"ON"}`), byte(1)).Return(nil)
	userID := 5
	historyRepo.On("AddSentCommand",
		mock.Anything,
		7,
		&userID,
		"state",
		"ON",
		"zigbee2mqtt/office-plug/set",
		`{"state":"ON"}`,
		10,
		mock.AnythingOfType("time.Time"),
	).Return(42, nil)

	result, err := svc.Send(context.Background(), 7, &gen.User{Id: userID, Permissions: []string{"control_sensors"}}, "state", "ON")

	require.NoError(t, err)
	assert.Equal(t, 42, result.ID)
	assert.Equal(t, "sent", result.Status)
}
