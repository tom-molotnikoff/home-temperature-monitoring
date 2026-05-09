//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	database "example/sensorHub/db"
	gen "example/sensorHub/gen"
	mqttpkg "example/sensorHub/mqtt"
	"example/sensorHub/service"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZigbee2MQTTBridgeDevices_BackfillsSensorMetadata(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	port := reserveTCPPort(t)
	sensorName := fmt.Sprintf("front-door-%d", port)

	embeddedBroker := mqttpkg.NewEmbeddedBroker(mqttpkg.BrokerConfig{
		TCPAddress: fmt.Sprintf(":%d", port),
	}, logger)
	require.NoError(t, embeddedBroker.Start())
	defer func() {
		require.NoError(t, embeddedBroker.Stop())
	}()

	sensorRepo := database.NewSensorRepository(env.DB, logger)
	readingsRepo := database.NewReadingsRepository(env.DB, logger)
	mtRepo := database.NewMeasurementTypeRepository(env.DB, logger)
	alertRepo := database.NewAlertRepository(env.DB, logger)
	brokerRepo := database.NewMQTTBrokerRepository(env.DB, logger)
	subRepo := database.NewMQTTSubscriptionRepository(env.DB, logger)

	sensorService := service.NewSensorService(sensorRepo, readingsRepo, mtRepo, alertRepo, nil, logger)
	connManager := mqttpkg.NewConnectionManager(sensorService, subRepo, brokerRepo, logger)

	brokerID, err := brokerRepo.Add(ctx, gen.MQTTBroker{
		Name:    fmt.Sprintf("integration-z2m-%d", port),
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

	defer func() {
		connManager.Stop()
		_ = subRepo.Delete(ctx, subID)
		_ = brokerRepo.Delete(ctx, brokerID)
		_ = client.DeleteSensor(sensorName)
	}()

	require.NoError(t, connManager.Start(ctx))
	require.Eventually(t, func() bool {
		return connManager.IsConnected(brokerID)
	}, 5*time.Second, 100*time.Millisecond)

	_, status := client.AddSensor(gen.Sensor{
		Name:         sensorName,
		SensorDriver: "mqtt-zigbee2mqtt",
		Config:       map[string]string{},
	})
	require.Equal(t, http.StatusCreated, status)

	mqttClient := pahomqtt.NewClient(
		pahomqtt.NewClientOptions().
			AddBroker(fmt.Sprintf("tcp://127.0.0.1:%d", port)).
			SetClientID(fmt.Sprintf("integration-z2m-publisher-%d", port)),
	)
	token := mqttClient.Connect()
	require.True(t, token.WaitTimeout(5*time.Second))
	require.NoError(t, token.Error())
	defer mqttClient.Disconnect(250)

	payload := fmt.Sprintf(`[
		{
			"ieee_address": "0x00158d00018255df",
			"friendly_name": %q,
			"definition": {
				"model": "MCCGQ11LM",
				"vendor": "Aqara",
				"description": "Door and window sensor",
				"exposes": [
					{"type": "binary", "property": "contact", "name": "contact", "access": 1}
				]
			}
		}
	]`, sensorName)

	pub := mqttClient.Publish("zigbee2mqtt/bridge/devices", 0, false, payload)
	require.True(t, pub.WaitTimeout(5*time.Second))
	require.NoError(t, pub.Error())

	var sensor gen.Sensor
	require.Eventually(t, func() bool {
		var currentStatus int
		sensor, currentStatus = client.GetSensorByName(sensorName)
		if currentStatus != http.StatusOK || sensor.Metadata == nil {
			return false
		}
		metadata := *sensor.Metadata
		return metadata["manufacturer"] == "Aqara" &&
			metadata["model"] == "MCCGQ11LM" &&
			metadata["description"] == "Door and window sensor" &&
			metadata["ieee_address"] == "0x00158d00018255df"
	}, 5*time.Second, 100*time.Millisecond)

	exposesJSON, err := json.Marshal((*sensor.Metadata)["exposes"])
	require.NoError(t, err)
	assert.JSONEq(t, `[{"type":"binary","property":"contact","name":"contact","access":1}]`, string(exposesJSON))
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
