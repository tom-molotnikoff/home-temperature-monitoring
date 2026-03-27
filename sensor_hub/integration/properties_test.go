//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProperties_GetAll(t *testing.T) {
	resp, status := client.GetProperties()
	require.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, resp)
}

func TestProperties_SetAndGet(t *testing.T) {
	status := client.SetProperty("sensor.collection.interval", "600")
	require.Equal(t, http.StatusAccepted, status)

	// Restore original value
	defer client.SetProperty("sensor.collection.interval", "300")

	resp, status := client.GetProperties()
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(resp), "600")
}
