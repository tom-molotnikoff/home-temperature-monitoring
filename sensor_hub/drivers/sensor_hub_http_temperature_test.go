package drivers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensorHubHTTPTemperature_Metadata(t *testing.T) {
	d := &SensorHubHTTPTemperature{client: http.DefaultClient}

	assert.Equal(t, "sensor-hub-http-temperature", d.Type())
	assert.Equal(t, "Sensor Hub HTTP Temperature", d.DisplayName())
	assert.NotEmpty(t, d.Description())

	mt := d.SupportedMeasurementTypes()
	require.Len(t, mt, 1)
	assert.Equal(t, "temperature", mt[0].Name)
	assert.Equal(t, "numeric", mt[0].Category)

	cf := d.ConfigFields()
	require.Len(t, cf, 1)
	assert.Equal(t, "url", cf[0].Key)
	assert.True(t, cf[0].Required)
	assert.False(t, cf[0].Sensitive)
}

func TestSensorHubHTTPTemperature_CollectReadings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/temperature", r.URL.Path)
		json.NewEncoder(w).Encode(rawTempResponse{Temperature: 22.5, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	d := &SensorHubHTTPTemperature{client: server.Client()}
	sensor := types.Sensor{Name: "test-sensor", Config: map[string]string{"url": server.URL}}

	readings, err := d.CollectReadings(context.Background(), sensor)

	require.NoError(t, err)
	require.Len(t, readings, 1)
	assert.Equal(t, "test-sensor", readings[0].SensorName)
	assert.Equal(t, "temperature", readings[0].MeasurementType)
	assert.InDelta(t, 22.5, *readings[0].NumericValue, 0.001)
	assert.Equal(t, "°C", readings[0].Unit)
}

func TestSensorHubHTTPTemperature_CollectReadings_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := &SensorHubHTTPTemperature{client: server.Client()}
	sensor := types.Sensor{Name: "bad-sensor", Config: map[string]string{"url": server.URL}}

	_, err := d.CollectReadings(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-200")
}

func TestSensorHubHTTPTemperature_CollectReadings_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	d := &SensorHubHTTPTemperature{client: server.Client()}
	sensor := types.Sensor{Name: "json-fail", Config: map[string]string{"url": server.URL}}

	_, err := d.CollectReadings(context.Background(), sensor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decoding")
}

func TestSensorHubHTTPTemperature_ValidateSensor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rawTempResponse{Temperature: 20.0, Time: "2025-01-01 12:00:00"})
	}))
	defer server.Close()

	d := &SensorHubHTTPTemperature{client: server.Client()}
	sensor := types.Sensor{Name: "valid-sensor", Config: map[string]string{"url": server.URL}}

	err := d.ValidateSensor(context.Background(), sensor)
	assert.NoError(t, err)
}

func TestSensorHubHTTPTemperature_CollectReadings_MissingURL(t *testing.T) {
	d := &SensorHubHTTPTemperature{client: http.DefaultClient}
	sensor := types.Sensor{Name: "no-url", Config: map[string]string{}}
	_, err := d.CollectReadings(context.Background(), sensor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no 'url' in config")
}

func TestSensorHubHTTPTemperature_RegisteredInRegistry(t *testing.T) {
	// init() registers the driver, but other tests may call Reset().
	// Re-register to verify the mechanism works.
	Reset()
	Register(&SensorHubHTTPTemperature{client: http.DefaultClient})

	d, ok := Get("sensor-hub-http-temperature")
	assert.True(t, ok, "driver should be findable after registration")
	assert.Equal(t, "sensor-hub-http-temperature", d.Type())
}
