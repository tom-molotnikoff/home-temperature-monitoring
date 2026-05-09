//go:build integration

package integration

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	database "example/sensorHub/db"
	gen "example/sensorHub/gen"
	mqttpkg "example/sensorHub/mqtt"
	"example/sensorHub/testharness"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type commandFixture struct {
	port     int
	brokerID int
	subID    int
	sensor   gen.Sensor
	stop     func()
}

func TestSendSensorCommand_PublishesAndPersistsSentCommand(t *testing.T) {
	fixture := setupCommandFixture(t, fmt.Sprintf("office-plug-%d", reserveTCPPort(t)))
	defer fixture.stop()

	subscriber := pahomqtt.NewClient(
		pahomqtt.NewClientOptions().
			AddBroker(fmt.Sprintf("tcp://127.0.0.1:%d", fixture.port)).
			SetClientID(fmt.Sprintf("integration-command-subscriber-%d", fixture.port)),
	)
	token := subscriber.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())
	defer subscriber.Disconnect(250)

	messageCh := make(chan pahomqtt.Message, 1)
	token = subscriber.Subscribe(fmt.Sprintf("zigbee2mqtt/%s/set", fixture.sensor.Name), 1, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		messageCh <- msg
	})
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())

	result, status := client.SendSensorCommand(fixture.sensor.Id, "state", "ON")
	require.Equal(t, http.StatusAccepted, status)
	assert.Equal(t, "state", result.Property)
	assert.Equal(t, "ON", result.Value)
	assert.Equal(t, gen.Sent, result.Status)
	assert.NotZero(t, result.Id)

	select {
	case msg := <-messageCh:
		assert.Equal(t, fmt.Sprintf("zigbee2mqtt/%s/set", fixture.sensor.Name), msg.Topic())
		assert.JSONEq(t, `{"state":"ON"}`, string(msg.Payload()))
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for published command")
	}

	adminUserID := lookupUserID(t, env.AdminUser)

	var userID int
	var property, value, statusValue, mqttTopic, mqttPayload string
	require.NoError(t, env.DB.QueryRow(`
		SELECT user_id, property, value, status, mqtt_topic, mqtt_payload
		FROM sensor_command_history
		WHERE id = ?
	`, result.Id).Scan(&userID, &property, &value, &statusValue, &mqttTopic, &mqttPayload))

	assert.Equal(t, adminUserID, userID)
	assert.Equal(t, "state", property)
	assert.Equal(t, "ON", value)
	assert.Equal(t, "sent", statusValue)
	assert.Equal(t, fmt.Sprintf("zigbee2mqtt/%s/set", fixture.sensor.Name), mqttTopic)
	assert.JSONEq(t, `{"state":"ON"}`, mqttPayload)
}

func TestSendSensorCommand_ViewerGetsForbidden(t *testing.T) {
	fixture := setupCommandFixture(t, fmt.Sprintf("viewer-plug-%d", reserveTCPPort(t)))
	defer fixture.stop()

	_, status := client.CreateUser(gen.CreateUserRequest{
		Username: fmt.Sprintf("command-viewer-%d", fixture.port),
		Password: "viewerpass123",
		Email:    ptrStr(fmt.Sprintf("command-viewer-%d@test.com", fixture.port)),
		Roles:    &[]string{"viewer"},
	})
	require.Equal(t, http.StatusCreated, status)

	viewer := testharness.NewClient(t, env.ServerURL)
	require.Equal(t, http.StatusOK, viewer.Login(fmt.Sprintf("command-viewer-%d", fixture.port), "viewerpass123"))
	require.Equal(t, http.StatusOK, viewer.ChangePassword("viewerpass123"))

	_, status = viewer.SendSensorCommand(fixture.sensor.Id, "state", "ON")
	assert.Equal(t, http.StatusForbidden, status)
}

func TestSendSensorCommand_BrokerDisconnectedReturnsServiceUnavailable(t *testing.T) {
	fixture := setupCommandFixture(t, fmt.Sprintf("disconnected-plug-%d", reserveTCPPort(t)))
	defer fixture.stop()

	env.ConnectionManager.DisconnectBroker(fixture.brokerID)
	require.Eventually(t, func() bool {
		return !env.ConnectionManager.IsConnected(fixture.brokerID)
	}, 5*time.Second, 100*time.Millisecond)

	_, status := client.SendSensorCommand(fixture.sensor.Id, "state", "ON")
	assert.Equal(t, http.StatusServiceUnavailable, status)
}

func TestSendSensorCommand_DisabledSensorReturnsConflict(t *testing.T) {
	fixture := setupCommandFixture(t, fmt.Sprintf("disabled-plug-%d", reserveTCPPort(t)))
	defer fixture.stop()

	require.Equal(t, http.StatusOK, client.DisableSensor(fixture.sensor.Name))

	_, status := client.SendSensorCommand(fixture.sensor.Id, "state", "ON")
	assert.Equal(t, http.StatusConflict, status)
}

func setupCommandFixture(t *testing.T, sensorName string) commandFixture {
	t.Helper()

	ctx := context.Background()
	logger := slog.Default()
	port := reserveTCPPort(t)

	embeddedBroker := mqttpkg.NewEmbeddedBroker(mqttpkg.BrokerConfig{
		TCPAddress: fmt.Sprintf(":%d", port),
	}, logger)
	require.NoError(t, embeddedBroker.Start())

	brokerRepo := database.NewMQTTBrokerRepository(env.DB, logger)
	subRepo := database.NewMQTTSubscriptionRepository(env.DB, logger)
	sensorRepo := database.NewSensorRepository(env.DB, logger)

	brokerID, err := brokerRepo.Add(ctx, gen.MQTTBroker{
		Name:    fmt.Sprintf("integration-command-broker-%d", port),
		Host:    "127.0.0.1",
		Port:    port,
		Type:    "external",
		Enabled: true,
	})
	require.NoError(t, err)

	subID, err := subRepo.Add(ctx, gen.MQTTSubscription{
		BrokerId:     brokerID,
		TopicPattern: "zigbee2mqtt/#",
		DriverType:   "mqtt-zigbee2mqtt",
		Enabled:      true,
	})
	require.NoError(t, err)

	metadata := map[string]interface{}{
		"exposes": []interface{}{
			map[string]interface{}{
				"type":      "binary",
				"property":  "state",
				"access":    float64(7),
				"value_on":  "ON",
				"value_off": "OFF",
			},
		},
	}
	require.NoError(t, sensorRepo.AddSensor(ctx, gen.Sensor{
		Name:         sensorName,
		SensorDriver: "mqtt-zigbee2mqtt",
		Status:       gen.SensorStatusActive,
		Config:       map[string]string{},
		Metadata:     &metadata,
	}))

	sensor, err := sensorRepo.GetSensorByName(ctx, sensorName)
	require.NoError(t, err)
	require.NotNil(t, sensor)

	broker := gen.MQTTBroker{
		Id:      &brokerID,
		Name:    fmt.Sprintf("integration-command-broker-%d", port),
		Host:    "127.0.0.1",
		Port:    port,
		Type:    "external",
		Enabled: true,
	}
	require.NoError(t, env.ConnectionManager.ConnectBroker(ctx, broker))
	require.Eventually(t, func() bool {
		return env.ConnectionManager.IsConnected(brokerID)
	}, 5*time.Second, 100*time.Millisecond)

	stop := func() {
		env.ConnectionManager.DisconnectBroker(brokerID)
		_ = sensorRepo.DeleteSensorByName(ctx, sensorName)
		_ = subRepo.Delete(ctx, subID)
		_ = brokerRepo.Delete(ctx, brokerID)
		require.NoError(t, embeddedBroker.Stop())
	}

	return commandFixture{
		port:     port,
		brokerID: brokerID,
		subID:    subID,
		sensor:   *sensor,
		stop:     stop,
	}
}

func lookupUserID(t *testing.T, username string) int {
	t.Helper()

	var id int
	require.NoError(t, env.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id))
	return id
}
