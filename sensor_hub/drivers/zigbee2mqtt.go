package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"example/sensorHub/types"
)

func init() {
	Register(&Zigbee2MQTTDriver{})
}

// fieldMapping maps a Zigbee2MQTT JSON field to a measurement type definition.
type fieldMapping struct {
	MeasurementType string
	DisplayName     string
	Unit            string
	Category        string // "numeric" or "binary"
}

// knownFields maps Zigbee2MQTT JSON field names to their measurement type
// definitions. Zigbee2MQTT normalises field names across hardware vendors,
// so "temperature" is always "temperature" regardless of the device model.
var knownFields = map[string]fieldMapping{
	"temperature":       {MeasurementType: "temperature", DisplayName: "Temperature", Unit: "°C", Category: "numeric"},
	"humidity":          {MeasurementType: "humidity", DisplayName: "Humidity", Unit: "%", Category: "numeric"},
	"pressure":          {MeasurementType: "pressure", DisplayName: "Pressure", Unit: "hPa", Category: "numeric"},
	"battery":           {MeasurementType: "battery", DisplayName: "Battery", Unit: "%", Category: "numeric"},
	"voltage":           {MeasurementType: "voltage", DisplayName: "Voltage", Unit: "mV", Category: "numeric"},
	"linkquality":       {MeasurementType: "link_quality", DisplayName: "Link Quality", Unit: "lqi", Category: "numeric"},
	"illuminance":       {MeasurementType: "illuminance", DisplayName: "Illuminance", Unit: "lx", Category: "numeric"},
	"illuminance_lux":   {MeasurementType: "illuminance", DisplayName: "Illuminance", Unit: "lx", Category: "numeric"},
	"power":             {MeasurementType: "power", DisplayName: "Power", Unit: "W", Category: "numeric"},
	"energy":            {MeasurementType: "energy", DisplayName: "Energy", Unit: "kWh", Category: "numeric"},
	"current":           {MeasurementType: "current", DisplayName: "Current", Unit: "A", Category: "numeric"},
	"co2":               {MeasurementType: "co2", DisplayName: "CO₂", Unit: "ppm", Category: "numeric"},
	"voc":               {MeasurementType: "voc", DisplayName: "VOC", Unit: "ppb", Category: "numeric"},
	"formaldehyde":      {MeasurementType: "formaldehyde", DisplayName: "Formaldehyde", Unit: "mg/m³", Category: "numeric"},
	"pm25":              {MeasurementType: "pm25", DisplayName: "PM2.5", Unit: "µg/m³", Category: "numeric"},
	"soil_moisture":     {MeasurementType: "soil_moisture", DisplayName: "Soil Moisture", Unit: "%", Category: "numeric"},
	"occupancy":         {MeasurementType: "occupancy", DisplayName: "Occupancy", Unit: "", Category: "binary"},
	"contact":           {MeasurementType: "contact", DisplayName: "Contact", Unit: "", Category: "binary"},
	"water_leak":        {MeasurementType: "water_leak", DisplayName: "Water Leak", Unit: "", Category: "binary"},
	"smoke":             {MeasurementType: "smoke", DisplayName: "Smoke", Unit: "", Category: "binary"},
	"carbon_monoxide":   {MeasurementType: "carbon_monoxide", DisplayName: "Carbon Monoxide", Unit: "", Category: "binary"},
	"tamper":            {MeasurementType: "tamper", DisplayName: "Tamper", Unit: "", Category: "binary"},
	"vibration":         {MeasurementType: "vibration", DisplayName: "Vibration", Unit: "", Category: "binary"},
	"state":             {MeasurementType: "state", DisplayName: "State", Unit: "", Category: "binary"},
}

// Zigbee2MQTTDriver parses MQTT messages from a Zigbee2MQTT bridge.
//
// Zigbee2MQTT publishes to topics like: zigbee2mqtt/<device_friendly_name>
// Each message is a flat JSON object with normalised field names.
type Zigbee2MQTTDriver struct{}

var _ PushDriver = (*Zigbee2MQTTDriver)(nil)

func (d *Zigbee2MQTTDriver) Type() string        { return "mqtt-zigbee2mqtt" }
func (d *Zigbee2MQTTDriver) DisplayName() string  { return "Zigbee2MQTT" }
func (d *Zigbee2MQTTDriver) Description() string {
	return "Zigbee devices via the Zigbee2MQTT bridge — temperature, humidity, contact sensors, smart plugs, and more"
}

func (d *Zigbee2MQTTDriver) ConfigFields() []ConfigFieldSpec {
	return nil // Push drivers have no sensor-level config
}

func (d *Zigbee2MQTTDriver) SupportedMeasurementTypes() []types.MeasurementType {
	var mts []types.MeasurementType
	seen := make(map[string]bool)
	for _, fm := range knownFields {
		if seen[fm.MeasurementType] {
			continue
		}
		seen[fm.MeasurementType] = true
		mts = append(mts, types.MeasurementType{
			Name:        fm.MeasurementType,
			DisplayName: fm.DisplayName,
			Unit:        fm.Unit,
			Category:    fm.Category,
		})
	}
	return mts
}

func (d *Zigbee2MQTTDriver) ValidateSensor(_ context.Context, _ types.Sensor) error {
	return nil // Push-based sensors have nothing to validate locally
}

// IdentifyDevice extracts the device name from a Zigbee2MQTT topic.
// Topics follow the pattern: <base_topic>/<friendly_name>
// For example: zigbee2mqtt/living-room-sensor → "living-room-sensor"
//
// System topics (bridge/..., bridge/log, bridge/state) are ignored.
func (d *Zigbee2MQTTDriver) IdentifyDevice(topic string, _ []byte) (string, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid zigbee2mqtt topic: %s", topic)
	}

	deviceName := parts[len(parts)-1]

	// Skip system topics
	if len(parts) >= 2 && parts[len(parts)-2] == "bridge" {
		return "", fmt.Errorf("system topic, not a device: %s", topic)
	}
	if deviceName == "bridge" {
		return "", fmt.Errorf("system topic, not a device: %s", topic)
	}

	// Skip sub-paths like zigbee2mqtt/device/set or zigbee2mqtt/device/get
	if deviceName == "set" || deviceName == "get" || deviceName == "availability" {
		return "", fmt.Errorf("command/availability topic, not a device message: %s", topic)
	}

	// For topics with more than 2 parts (e.g., zigbee2mqtt/room/sensor),
	// use the last segment after the base topic prefix
	if len(parts) > 2 {
		deviceName = strings.Join(parts[1:], "/")
		// But strip trailing /set, /get, /availability
		for _, suffix := range []string{"/set", "/get", "/availability"} {
			deviceName = strings.TrimSuffix(deviceName, suffix)
		}
	}

	return deviceName, nil
}

// ParseMessage extracts readings from a Zigbee2MQTT JSON payload.
// It maps known JSON fields to typed readings and ignores unknown fields.
func (d *Zigbee2MQTTDriver) ParseMessage(topic string, payload []byte) ([]types.Reading, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %w", err)
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	var readings []types.Reading

	for key, val := range data {
		fm, ok := knownFields[key]
		if !ok {
			continue
		}

		reading := types.Reading{
			MeasurementType: fm.MeasurementType,
			Unit:            fm.Unit,
			Time:            now,
		}

		switch fm.Category {
		case "numeric":
			numVal, ok := toFloat64(val)
			if !ok {
				continue
			}
			reading.NumericValue = &numVal
		case "binary":
			textVal := toBoolString(val)
			reading.TextState = &textVal
		}

		readings = append(readings, reading)
	}

	return readings, nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func toBoolString(v interface{}) string {
	switch b := v.(type) {
	case bool:
		if b {
			return "true"
		}
		return "false"
	case string:
		lower := strings.ToLower(b)
		if lower == "on" || lower == "true" || lower == "1" {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}
