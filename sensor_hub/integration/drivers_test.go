//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"example/sensorHub/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrivers_List(t *testing.T) {
	body, status := client.GetJSON("/api/drivers")
	require.Equal(t, http.StatusOK, status)

	var result []map[string]interface{}
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result, "should have at least one driver")

	// The built-in HTTP temperature driver should be present
	found := false
	for _, d := range result {
		if d["type"] == "sensor-hub-http-temperature" {
			found = true
			assert.Equal(t, "Sensor Hub HTTP Temperature", d["display_name"])
			configFields, ok := d["config_fields"].([]interface{})
			require.True(t, ok)
			require.Len(t, configFields, 1)

			field := configFields[0].(map[string]interface{})
			assert.Equal(t, "url", field["key"])
			assert.Equal(t, true, field["required"])
			break
		}
	}
	assert.True(t, found, "sensor-hub-http-temperature driver should be listed")

	// Cross-check: the number of drivers returned should match the registry
	allDrivers := drivers.All()
	assert.Len(t, result, len(allDrivers))
}
