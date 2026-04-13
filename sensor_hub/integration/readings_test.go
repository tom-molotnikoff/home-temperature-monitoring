//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadings_BetweenDates(t *testing.T) {
	ensureSensorsRegistered(t)
	client.CollectAll()

	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour).Format("2006-01-02")
	to := now.Add(24 * time.Hour).Format("2006-01-02")

	readings, status := client.GetReadingsBetween(from, to, "")
	require.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, readings)
}

func TestReadings_FilterBySensor(t *testing.T) {
	ensureSensorsRegistered(t)
	client.CollectAll()

	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour).Format("2006-01-02")
	to := now.Add(24 * time.Hour).Format("2006-01-02")

	readings, status := client.GetReadingsBetween(from, to, "Mock Sensor 1")
	require.Equal(t, http.StatusOK, status)

	for _, r := range readings {
		assert.Equal(t, "Mock Sensor 1", r.SensorName)
	}
}

func TestReadings_FilterBySensorCaseInsensitive(t *testing.T) {
	ensureSensorsRegistered(t)
	client.CollectAll()

	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour).Format("2006-01-02")
	to := now.Add(24 * time.Hour).Format("2006-01-02")

	readings, status := client.GetReadingsBetween(from, to, "mock sensor 1")
	require.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, readings)
}

func TestReadings_NoResults(t *testing.T) {
	readings, status := client.GetReadingsBetween("2020-01-01", "2020-01-02", "")
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, readings)
}

func TestReadings_ISODatetimeRange(t *testing.T) {
	ensureSensorsRegistered(t)
	client.CollectAll()

	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour).Format(time.RFC3339)
	to := now.Add(1 * time.Hour).Format(time.RFC3339)

	readings, status := client.GetReadingsBetween(from, to, "")
	require.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, readings)
}

func TestReadings_DatetimeNarrowerThanDate(t *testing.T) {
	ensureSensorsRegistered(t)
	client.CollectAll()

	// Use a range far in the past — should return nothing
	readings, status := client.GetReadingsBetween("2020-06-15T10:00:00Z", "2020-06-15T11:00:00Z", "")
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, readings)
}
