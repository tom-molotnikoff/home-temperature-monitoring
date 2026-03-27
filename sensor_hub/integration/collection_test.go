//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ensureSensorsRegistered adds mock sensors if not already present.
func ensureSensorsRegistered(t *testing.T) {
	t.Helper()
	sensors, _ := client.GetAllSensors()
	registered := make(map[string]bool)
	for _, s := range sensors {
		registered[s.Name] = true
	}

	mockSensors := []types.Sensor{
		{Name: "Mock Sensor 1", Type: "Temperature", URL: mockSensorURLs[0]},
		{Name: "Mock Sensor 2", Type: "Temperature", URL: mockSensorURLs[1]},
	}

	for _, s := range mockSensors {
		if !registered[s.Name] {
			_, status := client.AddSensor(s)
			if status != http.StatusCreated {
				t.Logf("warning: failed to add sensor %s (status %d)", s.Name, status)
			}
		}
	}
}

func TestCollection_CollectAll(t *testing.T) {
	ensureSensorsRegistered(t)

	_, status := client.CollectAll()
	require.Equal(t, http.StatusOK, status)

	// Verify readings were stored
	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour).Format("2006-01-02")
	to := now.Add(24 * time.Hour).Format("2006-01-02")

	readings, status := client.GetReadingsBetween(from, to, "")
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, readings, "collect-all should have stored readings")

	for _, r := range readings {
		assert.GreaterOrEqual(t, r.Temperature, 18.0)
		assert.LessOrEqual(t, r.Temperature, 22.0)
		assert.NotEmpty(t, r.Time)
		assert.NotEmpty(t, r.SensorName)
	}
}

func TestCollection_CollectByName(t *testing.T) {
	ensureSensorsRegistered(t)

	_, status := client.CollectByName("Mock Sensor 1")
	require.Equal(t, http.StatusOK, status)
}

func TestCollection_CollectByNameCaseInsensitive(t *testing.T) {
	ensureSensorsRegistered(t)

	_, status := client.CollectByName("mock sensor 1")
	require.Equal(t, http.StatusOK, status)
}

func TestCollection_CollectAllWithDisabledSensor(t *testing.T) {
	ensureSensorsRegistered(t)

	client.DisableSensor("Mock Sensor 2")
	defer client.EnableSensor("Mock Sensor 2")

	_, status := client.CollectAll()
	require.Equal(t, http.StatusOK, status)
}
