package drivers

import (
	"context"
	"fmt"
	"sync"

	"example/sensorHub/types"
)

// SensorDriver defines the interface for a sensor device driver.
// Each sensor device type is a single Go file implementing this interface.
type SensorDriver interface {
	// Type returns the unique identifier for this driver (e.g. "sensor-hub-http-temperature").
	Type() string
	// DisplayName returns a human-readable name for this driver.
	DisplayName() string
	// Description returns a short description of the driver.
	Description() string
	// SupportedMeasurementTypes returns the measurement types this driver can produce.
	SupportedMeasurementTypes() []types.MeasurementType
	// CollectReadings fetches current readings from the given sensor.
	CollectReadings(ctx context.Context, sensor types.Sensor) ([]types.Reading, error)
	// ValidateSensor checks whether a sensor's configuration is valid for this driver.
	ValidateSensor(ctx context.Context, sensor types.Sensor) error
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
