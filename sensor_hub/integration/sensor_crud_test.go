//go:build integration

package integration

import (
	"net/http"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensor_AddAndList(t *testing.T) {
	sensor := types.Sensor{
		Name: "Integration Test Sensor",
		SensorDriver: "sensor-hub-http-temperature",
		URL:  mockSensorURLs[0],
	}
	_, status := client.AddSensor(sensor)
	require.Equal(t, http.StatusCreated, status)

	sensors, status := client.GetAllSensors()
	require.Equal(t, http.StatusOK, status)

	found := false
	for _, s := range sensors {
		if s.Name == "Integration Test Sensor" {
			found = true
			assert.True(t, s.Id > 0)
			break
		}
	}
	assert.True(t, found, "sensor should appear in list")
}

func TestSensor_GetByName(t *testing.T) {
	sensor, status := client.GetSensorByName("Integration Test Sensor")
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Integration Test Sensor", sensor.Name)
}

func TestSensor_GetByNameCaseInsensitive(t *testing.T) {
	sensor, status := client.GetSensorByName("integration test sensor")
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Integration Test Sensor", sensor.Name)
}

func TestSensor_DisableAndEnable(t *testing.T) {
	status := client.DisableSensor("Integration Test Sensor")
	assert.Equal(t, http.StatusOK, status)

	sensor, _ := client.GetSensorByName("Integration Test Sensor")
	assert.False(t, sensor.Enabled)

	status = client.EnableSensor("Integration Test Sensor")
	assert.Equal(t, http.StatusOK, status)

	sensor, _ = client.GetSensorByName("Integration Test Sensor")
	assert.True(t, sensor.Enabled)
}

func TestSensor_DeleteAndVerifyGone(t *testing.T) {
	sensor := types.Sensor{
		Name: "Temp Sensor To Delete",
		SensorDriver: "sensor-hub-http-temperature",
		URL:  mockSensorURLs[1],
	}
	_, status := client.AddSensor(sensor)
	require.Equal(t, http.StatusCreated, status)

	status = client.DeleteSensor("Temp Sensor To Delete")
	assert.Equal(t, http.StatusOK, status)

	_, status = client.GetSensorByName("Temp Sensor To Delete")
	assert.NotEqual(t, http.StatusOK, status)
}
