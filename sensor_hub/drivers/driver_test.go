package drivers

import (
	"context"
	"testing"

	"example/sensorHub/types"

	"github.com/stretchr/testify/assert"
)

type stubDriver struct {
	driverType string
}

func (s *stubDriver) Type() string                        { return s.driverType }
func (s *stubDriver) DisplayName() string                 { return "Stub" }
func (s *stubDriver) Description() string                 { return "A stub driver for testing" }
func (s *stubDriver) ConfigFields() []ConfigFieldSpec     { return nil }
func (s *stubDriver) SupportedMeasurementTypes() []types.MeasurementType { return nil }
func (s *stubDriver) ValidateSensor(_ context.Context, _ types.Sensor) error { return nil }

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
