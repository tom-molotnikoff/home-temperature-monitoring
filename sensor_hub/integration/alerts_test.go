//go:build integration

package integration

import (
	"net/http"
	"testing"

	"example/sensorHub/testharness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlerts_CreateAndGetRule(t *testing.T) {
	ensureSensorsRegistered(t)

	sensor, _ := client.GetSensorByName("Mock Sensor 1")
	require.True(t, sensor.Id > 0)

	rule := testharness.AlertRuleRequest{
		SensorID:          sensor.Id,
		MeasurementTypeId: 1, // temperature
		AlertType:         "numeric_range",
		HighThreshold:     30.0,
		LowThreshold:      10.0,
		RateLimitHours:    6,
		Enabled:           true,
	}

	_, status := client.CreateAlertRule(rule)
	require.Equal(t, http.StatusCreated, status)

	resp, status := client.GetAlertRulesBySensorID(sensor.Id)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(resp), "numeric_range")
}

func TestAlerts_CollectionTriggersAlertCheck(t *testing.T) {
	ensureSensorsRegistered(t)

	// Collect readings — exercises the alert evaluation code path.
	// Mock sensors return 18-22°C, so a threshold of 30 won't fire,
	// but this validates the NullSQLiteTime fix doesn't cause scan errors.
	_, status := client.CollectAll()
	assert.Equal(t, http.StatusOK, status)

	sensor, _ := client.GetSensorByName("Mock Sensor 1")
	resp, status := client.GetAlertHistory(sensor.Id)
	require.Equal(t, http.StatusOK, status)
	// History may be empty (no alerts fired), but the endpoint should work
	assert.NotNil(t, resp)
}
