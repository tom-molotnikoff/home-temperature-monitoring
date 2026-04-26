package drivers

import (
	"context"
	"fmt"
	"sync"

	"example/sensorHub/gen"
)

// ConfigFieldSpec describes a single configuration field that a driver expects.
type ConfigFieldSpec struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive"`
	Default     string `json:"default,omitempty"`
}

// SensorDriver defines the base interface for all sensor device drivers.
// Drivers are either PullDriver (poll-based) or PushDriver (event-driven).
type SensorDriver interface {
	// Type returns the unique identifier for this driver (e.g. "sensor-hub-http-temperature").
	Type() string
	// DisplayName returns a human-readable name for this driver.
	DisplayName() string
	// Description returns a short description of the driver.
	Description() string
	// ConfigFields returns the schema of configuration fields this driver expects.
	ConfigFields() []ConfigFieldSpec
	// SupportedMeasurementTypes returns the measurement types this driver can produce.
	SupportedMeasurementTypes() []gen.MeasurementType
	// ValidateSensor checks whether a sensor's configuration is valid for this driver.
	ValidateSensor(ctx context.Context, sensor gen.Sensor) error
}

// PullDriver is a poll-based sensor driver. The service calls CollectReadings
// on a periodic schedule to fetch current values from the sensor.
type PullDriver interface {
	SensorDriver
	// CollectReadings fetches current readings from the given sensor.
	CollectReadings(ctx context.Context, sensor gen.Sensor) ([]gen.Reading, error)
}

// PushDriver is an event-driven sensor driver for MQTT ecosystems.
// Messages arrive via shared MQTT clients; the driver only parses payloads.
type PushDriver interface {
	SensorDriver
	// ParseMessage extracts readings from an MQTT message payload.
	ParseMessage(topic string, payload []byte) ([]gen.Reading, error)
	// IdentifyDevice returns a suggested sensor name from an MQTT message,
	// used during auto-discovery of new devices.
	IdentifyDevice(topic string, payload []byte) (string, error)
}

var (
	mu       sync.RWMutex
	registry = make(map[string]SensorDriver)
)

// Register adds a driver to the global registry. Called from driver init() functions.
// Panics if a driver with the same type is already registered.
func Register(d SensorDriver) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[d.Type()]; exists {
		panic(fmt.Sprintf("sensor driver already registered: %s", d.Type()))
	}
	registry[d.Type()] = d
}

// Get returns a registered driver by type, or false if not found.
func Get(driverType string) (SensorDriver, bool) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := registry[driverType]
	return d, ok
}

// All returns a slice of all registered drivers.
func All() []SensorDriver {
	mu.RLock()
	defer mu.RUnlock()
	all := make([]SensorDriver, 0, len(registry))
	for _, d := range registry {
		all = append(all, d)
	}
	return all
}

// Reset clears the registry. Only for use in tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]SensorDriver)
}
