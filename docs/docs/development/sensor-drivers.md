# Sensor Drivers

Sensor Hub uses a **driver** model to support arbitrary sensor hardware and
protocols. Each driver is a single Go file that implements the `SensorDriver`
interface, self-registers at startup, and tells the system how to communicate
with a particular class of sensor.

## What is a Sensor Driver?

A sensor driver is the bridge between a physical (or virtual) sensor device and
the Sensor Hub data pipeline. It encapsulates:

- **Identity** — a unique type string, human-readable name, and description.
- **Measurement types** — which kinds of readings the sensor can produce
  (temperature, humidity, motion, etc.).
- **Data collection** — how to connect to the sensor and fetch readings.
- **Validation** — how to verify that a sensor's configuration is correct before
  the system starts collecting from it.

The rest of the application — storage, alerting, dashboards, WebSocket
broadcasts — is completely driver-agnostic. Once a driver returns a
`[]types.Reading`, the generic pipeline takes over.

## When should you write a new driver?

Write a new driver when you need to integrate a new **sensor communication
protocol or sensor device type**. Some examples:

| Scenario | Driver needed? |
|----------|---------------|
| Add support for Zigbee temperature/humidity sensors | ✅ Yes — new protocol |
| Add support for an MQTT-based weather station | ✅ Yes — new protocol |
| Add a second HTTP temperature sensor with a different JSON format | ✅ Yes — different payload format |
| Add a new measurement type (e.g. CO₂) to an existing driver | ❌ No — extend the existing driver |
| Change how alerts work for temperature readings | ❌ No — that's the alert service, not the driver |

## The SensorDriver Interface

Every driver must implement this interface, defined in `drivers/driver.go`:

```go
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
```

### Method responsibilities

#### `Type() string`

Returns the **unique, stable identifier** for this driver. This string is stored
in the database (`sensors.sensor_driver` column) and used in API requests. Use
kebab-case: `"vendor-protocol-device"`.

Choose a name that is specific enough to never collide with another driver but
general enough that firmware updates don't invalidate it. Good examples:

- `"sensor-hub-http-temperature"`
- `"zigbee-aqara-multisensor"`
- `"mqtt-tasmota-energy"`

#### `DisplayName() string`

A human-readable name shown in the UI and CLI. Example: `"Sensor Hub HTTP
Temperature"`.

#### `Description() string`

A short sentence for tooltips and documentation.

#### `SupportedMeasurementTypes() []types.MeasurementType`

Returns the list of measurement types this driver can produce. Each type has a
`Name`, `DisplayName`, `Unit`, and `Category`. The category is either `"numeric"`
(for continuous values like temperature, humidity) or `"binary"` (for on/off
states like motion, contact).

These measurement types must already exist in the `measurement_types` database
table (seeded by migration 000006). If your driver introduces a completely new
measurement type, add it via a new migration.

```go
func (d *MyDriver) SupportedMeasurementTypes() []types.MeasurementType {
    return []types.MeasurementType{
        {Name: "temperature", DisplayName: "Temperature", Unit: "°C", Category: "numeric"},
        {Name: "humidity", DisplayName: "Humidity", Unit: "%", Category: "numeric"},
    }
}
```

#### `CollectReadings(ctx, sensor) ([]Reading, error)`

The core method. Called by the sensor service on every collection cycle (and on
demand via the API). It must:

1. **Connect** to the sensor using `sensor.URL` (and any other fields on the
   `types.Sensor` struct).
2. **Fetch** the raw data from the device.
3. **Parse** the response into one or more `types.Reading` structs.
4. **Return** the readings, or an error if something went wrong.

Each `Reading` must set:

| Field | Required | Description |
|-------|----------|-------------|
| `SensorName` | ✅ | Must match `sensor.Name` |
| `MeasurementType` | ✅ | Must match a name from `SupportedMeasurementTypes()` |
| `NumericValue` | For `numeric` types | Pointer to float64 |
| `TextState` | For `binary` types | Pointer to string (e.g. `"open"`, `"closed"`) |
| `Unit` | ✅ | The unit string (e.g. `"°C"`, `"%"`, `"lx"`) |
| `Time` | ✅ | Timestamp in `"2006-01-02 15:04:05"` format |

For numeric measurement types, set `NumericValue` and leave `TextState` nil.
For binary measurement types, set `TextState` and leave `NumericValue` nil.

A multi-sensor driver (e.g. a weather station reporting temperature and humidity)
should return multiple readings in a single call.

**Error handling:** Return a descriptive, wrapped error. The sensor service will
mark the sensor health as `"bad"` and log the error. Do not panic.

**Context:** Respect the `ctx` parameter. Use it for HTTP requests
(`http.NewRequestWithContext`) and check for cancellation in long-running
operations.

#### `ValidateSensor(ctx, sensor) error`

Called when a sensor is added or updated, before it enters the collection cycle.
The simplest implementation is to call `CollectReadings` and check for errors:

```go
func (d *MyDriver) ValidateSensor(ctx context.Context, sensor types.Sensor) error {
    _, err := d.CollectReadings(ctx, sensor)
    return err
}
```

If your driver can do a cheaper "ping" or connectivity check, implement that
instead.

## Driver Lifecycle

Drivers participate in the application lifecycle at several points:

```
Application Start
    │
    ├─ 1. init() → drivers.Register(d)     — self-registration
    │
    ├─ 2. Sensor added via API
    │     └─ ValidateSensor()               — reachability check
    │
    ├─ 3. Periodic collection tick (every N seconds)
    │     ├─ Get all enabled sensors
    │     ├─ drivers.Get(sensor.SensorDriver) — look up the driver
    │     ├─ CollectReadings()               — fetch data
    │     ├─ Store readings in database
    │     ├─ Broadcast via WebSocket
    │     └─ Process alert rules
    │
    └─ 4. On-demand collection via API
          └─ Same as step 3, for a single sensor
```

The driver itself is **stateless** between calls. The system creates one instance
at startup (in `init()`) and reuses it for all sensors that reference that driver
type. Do not store per-sensor state in the driver struct — use `sensor.URL` and
other `types.Sensor` fields to distinguish between sensors.

## Writing a New Driver: Step by Step

### 1. Create the driver file

Create a new file in `sensor_hub/drivers/`. Name it after the protocol and
device, e.g. `mqtt_tasmota_energy.go`.

```go
package drivers

import (
    "context"
    "fmt"
    "example/sensorHub/types"
)

func init() {
    Register(&MQTTTasmotaEnergy{})
}

type MQTTTasmotaEnergy struct{}

func (d *MQTTTasmotaEnergy) Type() string        { return "mqtt-tasmota-energy" }
func (d *MQTTTasmotaEnergy) DisplayName() string  { return "Tasmota Energy (MQTT)" }
func (d *MQTTTasmotaEnergy) Description() string {
    return "Tasmota-flashed smart plug reporting power consumption via MQTT"
}

func (d *MQTTTasmotaEnergy) SupportedMeasurementTypes() []types.MeasurementType {
    return []types.MeasurementType{
        {Name: "power", DisplayName: "Power", Unit: "W", Category: "numeric"},
        {Name: "voltage", DisplayName: "Voltage", Unit: "V", Category: "numeric"},
    }
}

func (d *MQTTTasmotaEnergy) CollectReadings(ctx context.Context, sensor types.Sensor) ([]types.Reading, error) {
    // TODO: Connect to MQTT broker, read the latest values for sensor.URL (topic).
    return nil, fmt.Errorf("not yet implemented")
}

func (d *MQTTTasmotaEnergy) ValidateSensor(ctx context.Context, sensor types.Sensor) error {
    _, err := d.CollectReadings(ctx, sensor)
    return err
}
```

### 2. Write tests

Create a corresponding test file `mqtt_tasmota_energy_test.go`:

```go
package drivers

import (
    "context"
    "testing"

    "example/sensorHub/types"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMQTTTasmotaEnergy_Metadata(t *testing.T) {
    d := &MQTTTasmotaEnergy{}

    assert.Equal(t, "mqtt-tasmota-energy", d.Type())
    assert.Equal(t, "Tasmota Energy (MQTT)", d.DisplayName())
    assert.NotEmpty(t, d.Description())

    mt := d.SupportedMeasurementTypes()
    require.Len(t, mt, 2)
    assert.Equal(t, "power", mt[0].Name)
    assert.Equal(t, "numeric", mt[0].Category)
}

func TestMQTTTasmotaEnergy_CollectReadings_Success(t *testing.T) {
    d := &MQTTTasmotaEnergy{}
    sensor := types.Sensor{Name: "smart-plug-1", URL: "mqtt://broker:1883/tasmota/plug1"}

    readings, err := d.CollectReadings(context.Background(), sensor)

    require.NoError(t, err)
    require.Len(t, readings, 2)
    assert.Equal(t, "smart-plug-1", readings[0].SensorName)
    assert.Equal(t, "power", readings[0].MeasurementType)
}
```

**Test strategy:** Use `httptest.NewServer` for HTTP-based drivers, mock MQTT
brokers for MQTT drivers, or test data files for file-based drivers. Always test:

- Metadata (type, display name, supported measurement types)
- Successful collection with realistic data
- Error cases (network failure, invalid response, non-200 status)
- Validation behaviour
- Registration in the global registry

### 3. Add any new measurement types

If your driver introduces measurement types not already seeded, create a new
migration:

```sql
-- db/migrations/000007_add_energy_measurement_types.up.sql
INSERT OR IGNORE INTO measurement_types (name, display_name, category, default_unit) VALUES
    ('energy', 'Energy', 'numeric', 'kWh');
```

The currently seeded measurement types are: `temperature`, `humidity`,
`pressure`, `power`, `battery`, `voltage`, `luminance`, `motion`, `contact`,
`doorbell`.

### 4. Register via blank import

Drivers self-register in their `init()` function. The server binary activates
all drivers via a blank import in `cmd/serve.go`:

```go
import (
    _ "example/sensorHub/drivers" // register sensor drivers
)
```

This import is already present — you do not need to change it. Adding a new
`.go` file to the `drivers` package automatically includes its `init()` function.

### 5. No other code changes needed

That's the key advantage of the driver architecture. You do **not** need to
modify:

- The sensor service (it dispatches via `drivers.Get(sensor.SensorDriver)`)
- The readings repository (it stores any `[]types.Reading`)
- The API layer (it works with the generic `Reading` type)
- The alert system (it matches on `measurement_type` + `sensor_id`)
- The frontend (it renders readings generically by measurement type)

A sensor using your new driver is created through the normal sensor API with the
driver's `Type()` string as the `sensor_driver` field:

```json
{
    "name": "living-room-plug",
    "sensor_driver": "mqtt-tasmota-energy",
    "url": "mqtt://broker:1883/tasmota/plug1"
}
```

## Existing Drivers

### `sensor-hub-http-temperature`

| Property | Value |
|----------|-------|
| File | `drivers/sensor_hub_http_temperature.go` |
| Protocol | HTTP GET |
| Measurement types | `temperature` (numeric, °C) |
| Sensor URL format | `http://host:port` (appends `/temperature`) |
| Response format | `{"temperature": 22.5, "time": "2025-01-01 12:00:00"}` |

This is the built-in driver for the Sensor Hub's own ESP32-based temperature
sensors running the companion firmware. It makes a GET request to
`{sensor.URL}/temperature` and expects a JSON response with `temperature` (float)
and `time` (string) fields.

## Driver Registry

The global driver registry (`drivers/driver.go`) provides three functions:

| Function | Description |
|----------|-------------|
| `Register(d SensorDriver)` | Add a driver. Panics on duplicate type. Called from `init()`. |
| `Get(driverType string) (SensorDriver, bool)` | Look up a driver by type string. |
| `All() []SensorDriver` | Return all registered drivers. |

The registry is thread-safe (protected by `sync.RWMutex`). A `Reset()` function
exists for test isolation only — never call it in production code.

## Future: Other Driver Interfaces

The `SensorDriver` interface is specifically for **sensors** — devices that
produce readings. The naming is intentional: as Sensor Hub evolves, other driver
interfaces may be introduced for different device categories:

- **`ActuatorDriver`** — for switches, relays, and controllers that accept
  commands (on/off, set level, etc.)
- **`CameraDriver`** — for devices that produce image or video streams

Each would follow the same pattern: interface definition, global registry,
self-registration via `init()`, and a dedicated package for implementations.
