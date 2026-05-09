package drivers

import (
	"context"
	"encoding/json"
	"testing"

	gen "example/sensorHub/gen"

	"github.com/stretchr/testify/assert"
)

type stubDriver struct {
	driverType string
}

func (s *stubDriver) Type() string                                         { return s.driverType }
func (s *stubDriver) DisplayName() string                                  { return "Stub" }
func (s *stubDriver) Description() string                                  { return "A stub driver for testing" }
func (s *stubDriver) ConfigFields() []ConfigFieldSpec                      { return nil }
func (s *stubDriver) SupportedMeasurementTypes() []gen.MeasurementType     { return nil }
func (s *stubDriver) ValidateSensor(_ context.Context, _ gen.Sensor) error { return nil }

type stubCommandDriver struct {
	stubDriver
}

func (s *stubCommandDriver) ParseCapabilities(_ json.RawMessage) []gen.Capability {
	return []gen.Capability{{Property: "state", Type: gen.CapabilityTypeBinary}}
}

func (s *stubCommandDriver) BuildCommand(_ gen.Sensor, _ string, _ string) (string, []byte, error) {
	return "topic", []byte(`{"state":"ON"}`), nil
}

func TestRegister_And_Get(t *testing.T) {
	Reset()
	defer Reset()

	d := &stubDriver{driverType: "test-driver"}
	Register(d)

	got, ok := Get("test-driver")
	assert.True(t, ok)
	assert.Equal(t, "test-driver", got.Type())

	_, ok = Get("nonexistent")
	assert.False(t, ok)
}

func TestRegister_Duplicate_Panics(t *testing.T) {
	Reset()
	defer Reset()

	d := &stubDriver{driverType: "dup-driver"}
	Register(d)

	assert.Panics(t, func() {
		Register(d)
	})
}

func TestAll(t *testing.T) {
	Reset()
	defer Reset()

	Register(&stubDriver{driverType: "driver-a"})
	Register(&stubDriver{driverType: "driver-b"})

	all := All()
	assert.Len(t, all, 2)
}

func TestGetCommandDriver(t *testing.T) {
	Reset()
	defer Reset()

	Register(&stubDriver{driverType: "plain-driver"})
	Register(&stubCommandDriver{stubDriver: stubDriver{driverType: "command-driver"}})

	commandDriver, ok := GetCommandDriver("command-driver")
	assert.True(t, ok)
	assert.Equal(t, "command-driver", commandDriver.Type())

	_, ok = GetCommandDriver("plain-driver")
	assert.False(t, ok)

	_, ok = GetCommandDriver("missing-driver")
	assert.False(t, ok)
}
