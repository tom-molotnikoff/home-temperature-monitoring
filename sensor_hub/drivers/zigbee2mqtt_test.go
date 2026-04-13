package drivers

import (
	"context"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZigbee2MQTT_Type(t *testing.T) {
	d := &Zigbee2MQTTDriver{}
	assert.Equal(t, "mqtt-zigbee2mqtt", d.Type())
}

func TestZigbee2MQTT_IdentifyDevice_Standard(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	name, err := d.IdentifyDevice("zigbee2mqtt/living-room-sensor", nil)
	require.NoError(t, err)
	assert.Equal(t, "living-room-sensor", name)
}

func TestZigbee2MQTT_IdentifyDevice_MultiSegment(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	name, err := d.IdentifyDevice("zigbee2mqtt/room/sensor", nil)
	require.NoError(t, err)
	assert.Equal(t, "room/sensor", name)
}

func TestZigbee2MQTT_IdentifyDevice_SkipBridge(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	// Direct bridge topics
	_, err := d.IdentifyDevice("zigbee2mqtt/bridge", nil)
	assert.Error(t, err)

	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/state", nil)
	assert.Error(t, err)

	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/log", nil)
	assert.Error(t, err)

	// Nested bridge topics (these slipped through before the fix)
	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/response/permit_join", nil)
	assert.Error(t, err, "nested bridge/response topic should be filtered")

	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/response/device/rename", nil)
	assert.Error(t, err, "deeply nested bridge topic should be filtered")

	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/event", nil)
	assert.Error(t, err, "bridge/event should be filtered")

	_, err = d.IdentifyDevice("zigbee2mqtt/bridge/response/options", nil)
	assert.Error(t, err, "bridge/response/options should be filtered")
}

func TestZigbee2MQTT_IdentifyDevice_SkipCommandTopics(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	_, err := d.IdentifyDevice("zigbee2mqtt/device/set", nil)
	assert.Error(t, err)

	_, err = d.IdentifyDevice("zigbee2mqtt/device/get", nil)
	assert.Error(t, err)

	_, err = d.IdentifyDevice("zigbee2mqtt/device/availability", nil)
	assert.Error(t, err)
}

func TestZigbee2MQTT_IdentifyDevice_InvalidTopic(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	_, err := d.IdentifyDevice("singlesegment", nil)
	assert.Error(t, err)
}

func TestZigbee2MQTT_ParseMessage_TemperatureHumidity(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"temperature": 22.5, "humidity": 45.2, "battery": 87, "linkquality": 120}`
	readings, err := d.ParseMessage("zigbee2mqtt/sensor-1", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 4)

	readingMap := make(map[string]float64)
	for _, r := range readings {
		if r.NumericValue != nil {
			readingMap[r.MeasurementType] = *r.NumericValue
		}
	}

	assert.InDelta(t, 22.5, readingMap["temperature"], 0.001)
	assert.InDelta(t, 45.2, readingMap["humidity"], 0.001)
	assert.InDelta(t, 87.0, readingMap["battery"], 0.001)
	assert.InDelta(t, 120.0, readingMap["link_quality"], 0.001)
}

func TestZigbee2MQTT_ParseMessage_BinarySensor(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"occupancy": true, "battery": 100, "linkquality": 80}`
	readings, err := d.ParseMessage("zigbee2mqtt/motion-1", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 3)

	for _, r := range readings {
		if r.MeasurementType == "occupancy" {
			require.NotNil(t, r.TextState)
			assert.Equal(t, "true", *r.TextState)
		}
	}
}

func TestZigbee2MQTT_ParseMessage_ContactSensor(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"contact": false, "battery": 95}`
	readings, err := d.ParseMessage("zigbee2mqtt/door-1", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 2)

	for _, r := range readings {
		if r.MeasurementType == "contact" {
			require.NotNil(t, r.TextState)
			assert.Equal(t, "false", *r.TextState)
		}
	}
}

func TestZigbee2MQTT_ParseMessage_BatteryLow(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"battery": 100, "battery_low": false, "contact": true, "linkquality": 204}`
	readings, err := d.ParseMessage("zigbee2mqtt/back-door-contact", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 4)

	for _, r := range readings {
		if r.MeasurementType == "battery_low" {
			require.NotNil(t, r.TextState)
			assert.Equal(t, "false", *r.TextState)
			assert.Equal(t, "", r.Unit)
		}
	}
}

func TestZigbee2MQTT_ParseMessage_UnknownFieldsIgnored(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"temperature": 20.0, "unknown_field": "value", "another": 42}`
	readings, err := d.ParseMessage("zigbee2mqtt/sensor", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 1)
	assert.Equal(t, "temperature", readings[0].MeasurementType)
}

func TestZigbee2MQTT_ParseMessage_InvalidJSON(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	_, err := d.ParseMessage("zigbee2mqtt/sensor", []byte(`not json`))
	assert.Error(t, err)
}

func TestZigbee2MQTT_ParseMessage_EmptyPayload(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	readings, err := d.ParseMessage("zigbee2mqtt/sensor", []byte(`{}`))
	require.NoError(t, err)
	assert.Empty(t, readings)
}

func TestZigbee2MQTT_ParseMessage_StringBoolValues(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"state": "ON"}`
	readings, err := d.ParseMessage("zigbee2mqtt/plug-1", []byte(payload))

	require.NoError(t, err)
	assert.Len(t, readings, 1)
	assert.Equal(t, "state", readings[0].MeasurementType)
	require.NotNil(t, readings[0].TextState)
	assert.Equal(t, "true", *readings[0].TextState)
}

func TestZigbee2MQTT_ParseMessage_SmartPlug(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	payload := `{"current":0.04,"energy":1.23,"energy_month":0.5,"energy_today":0.1,"energy_yesterday":0.4,"linkquality":255,"power":3.91,"state":"ON","voltage":231.15}`
	readings, err := d.ParseMessage("zigbee2mqtt/office-plug", []byte(payload))

	require.NoError(t, err)

	readingMap := make(map[string]types.Reading)
	for _, r := range readings {
		readingMap[r.MeasurementType] = r
	}

	// Numeric fields
	assert.InDelta(t, 0.04, *readingMap["current"].NumericValue, 0.001)
	assert.InDelta(t, 1.23, *readingMap["energy"].NumericValue, 0.001)
	assert.InDelta(t, 0.5, *readingMap["energy_month"].NumericValue, 0.001)
	assert.InDelta(t, 0.1, *readingMap["energy_today"].NumericValue, 0.001)
	assert.InDelta(t, 0.4, *readingMap["energy_yesterday"].NumericValue, 0.001)
	assert.InDelta(t, 3.91, *readingMap["power"].NumericValue, 0.001)
	assert.InDelta(t, 255.0, *readingMap["link_quality"].NumericValue, 0.001)

	// Mains voltage stays as-is (231.15 V < 500 threshold)
	assert.InDelta(t, 231.15, *readingMap["voltage"].NumericValue, 0.001)
	assert.Equal(t, "V", readingMap["voltage"].Unit)

	// State is binary
	assert.Equal(t, "true", *readingMap["state"].TextState)
}

func TestZigbee2MQTT_ParseMessage_VoltageConversion(t *testing.T) {
	d := &Zigbee2MQTTDriver{}

	// Battery voltage in mV (>500) should be converted to V
	payload := `{"voltage": 3100, "battery": 95}`
	readings, err := d.ParseMessage("zigbee2mqtt/contact-sensor", []byte(payload))

	require.NoError(t, err)

	for _, r := range readings {
		if r.MeasurementType == "voltage" {
			require.NotNil(t, r.NumericValue)
			assert.InDelta(t, 3.1, *r.NumericValue, 0.001, "3100 mV should be converted to 3.1 V")
			assert.Equal(t, "V", r.Unit)
		}
	}

	// Mains voltage in V (<500) should stay as-is
	payload = `{"voltage": 231.72, "power": 0}`
	readings, err = d.ParseMessage("zigbee2mqtt/plug", []byte(payload))

	require.NoError(t, err)

	for _, r := range readings {
		if r.MeasurementType == "voltage" {
			require.NotNil(t, r.NumericValue)
			assert.InDelta(t, 231.72, *r.NumericValue, 0.001, "mains voltage should not be converted")
			assert.Equal(t, "V", r.Unit)
		}
	}
}

func TestZigbee2MQTT_ValidateSensor(t *testing.T) {
	d := &Zigbee2MQTTDriver{}
	err := d.ValidateSensor(context.Background(), types.Sensor{})
	assert.NoError(t, err)
}

func TestZigbee2MQTT_SupportedMeasurementTypes(t *testing.T) {
	d := &Zigbee2MQTTDriver{}
	mts := d.SupportedMeasurementTypes()
	assert.NotEmpty(t, mts)

	names := make(map[string]bool)
	for _, mt := range mts {
		names[mt.Name] = true
	}
	assert.True(t, names["temperature"])
	assert.True(t, names["humidity"])
	assert.True(t, names["occupancy"])
}

func TestZigbee2MQTT_ImplementsPushDriver(t *testing.T) {
	var drv SensorDriver = &Zigbee2MQTTDriver{}
	_, ok := drv.(PushDriver)
	assert.True(t, ok, "Zigbee2MQTTDriver should implement PushDriver")
}
