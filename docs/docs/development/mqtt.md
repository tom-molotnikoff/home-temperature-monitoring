# MQTT Ingest

This guide explains how MQTT-based sensor data flows through Sensor Hub, how to
configure brokers and subscriptions, and how to write new PushDrivers for
additional MQTT ecosystems.

## Overview

Sensor Hub supports two models for collecting sensor data:

| Model | Interface | Example | Data flow |
|-------|-----------|---------|-----------|
| **Pull** | `PullDriver` | HTTP temperature sensor | Service polls sensor on a timer |
| **Push** | `PushDriver` | Zigbee2MQTT, rtl_433 | MQTT messages arrive continuously |

Both models feed into the same downstream pipeline — readings are stored,
alerts are evaluated, and WebSocket broadcasts fire identically regardless of
how the data was collected.

## Architecture

```
┌────────────────────┐      ┌────────────────────┐
│  Zigbee2MQTT       │      │  Other MQTT source  │
│  (external broker) │      │  (external broker)  │
└────────┬───────────┘      └────────┬────────────┘
         │ MQTT                      │ MQTT
         ▼                           ▼
┌─────────────────────────────────────────────────┐
│              Connection Manager                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Client 1 │  │ Client 2 │  │ Client N │      │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘      │
│       └──────────────┼────────────┘              │
│                      ▼                           │
│           Route message to PushDriver            │
│           via subscription config                │
└──────────────────────┬──────────────────────────┘
                       │
              ┌────────▼────────┐
              │   PushDriver    │
              │  ParseMessage() │
              │  IdentifyDevice │
              └────────┬────────┘
                       │ []Reading
              ┌────────▼────────┐
              │  SensorService  │
              │  (same pipeline │
              │   as PullDriver)│
              └─────────────────┘
```

### Key packages

| Package | Purpose |
|---------|---------|
| `mqtt/` | Embedded broker (`broker.go`) and connection manager (`connection_manager.go`) |
| `drivers/` | PushDriver implementations (e.g. `zigbee2mqtt.go`) |
| `db/` | `mqtt_broker_repository.go`, `mqtt_subscription_repository.go` |
| `service/` | `mqtt_service.go` — broker/subscription CRUD and validation |
| `api/` | `mqtt_api.go` / `mqtt_routes.go` — REST endpoints |
| `cmd/` | `mqtt.go` — CLI commands |

## Data Flow: MQTT Message → Stored Reading

```
1. ConnectionManager.Start() connects to all enabled brokers
   and subscribes to topic patterns from mqtt_subscriptions table.

2. MQTT message arrives on topic matching a subscription's topic_pattern.

3. Connection manager looks up the subscription → gets driver_type.

4. drivers.Get(driverType) → PushDriver instance.

5. PushDriver.IdentifyDevice(topic, payload) → device name.

6. If device is unknown: auto-create sensor with status="pending".

7. PushDriver.ParseMessage(topic, payload) → []Reading.

8. SensorService.ServiceProcessPushReadings(ctx, sensorName, readings):
   a. Store readings in database
   b. Update sensor health status
   c. Evaluate alert rules
   d. Broadcast via WebSocket
```

## Configuration

### Embedded Broker

Sensor Hub can optionally run an embedded MQTT broker (mochi-mqtt) inside the
same process. This is useful for simple setups where you don't want to run
Mosquitto or another external broker.

Configured via `application.properties`:

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `mqtt.broker.enabled` | bool | `false` | Start the embedded MQTT broker |
| `mqtt.broker.port` | int | `1883` | TCP port for the embedded broker |

The embedded broker starts before the database is initialised and stops during
graceful shutdown (via defer in `serve.go`).

### Broker Records (Database)

All MQTT broker connections (both embedded and external) are configured in the
`mqtt_brokers` database table and managed via the API/CLI. Even if you enable
the embedded broker, you still need to create a broker record pointing to
`localhost:1883` so the connection manager knows to connect to it.

```sql
-- mqtt_brokers table
id          INTEGER PRIMARY KEY
name        TEXT NOT NULL UNIQUE
type        TEXT NOT NULL DEFAULT 'external'   -- 'embedded' or 'external'
host        TEXT NOT NULL
port        INTEGER NOT NULL DEFAULT 1883
username    TEXT
password    TEXT
use_tls     INTEGER NOT NULL DEFAULT 0
enabled     INTEGER NOT NULL DEFAULT 1
created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
```

### Subscription Records (Database)

Subscriptions link a broker to an MQTT topic pattern and a PushDriver:

```sql
-- mqtt_subscriptions table
id            INTEGER PRIMARY KEY
broker_id     INTEGER NOT NULL REFERENCES mqtt_brokers(id) ON DELETE CASCADE
topic_pattern TEXT NOT NULL            -- e.g. "zigbee2mqtt/#"
driver_type   TEXT NOT NULL            -- e.g. "mqtt-zigbee2mqtt"
qos           INTEGER NOT NULL DEFAULT 0
enabled       INTEGER NOT NULL DEFAULT 1
created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
```

The `driver_type` must match the `Type()` return value of a registered
PushDriver.

### Sensor Status

Sensors have a `status` field that supports auto-discovery:

| Status | Meaning |
|--------|---------|
| `active` | Normal sensor, readings are collected and stored |
| `pending` | Auto-discovered via MQTT, awaiting user approval |
| `dismissed` | User has dismissed this device (can be restored later) |

When a PushDriver identifies a device name that doesn't match any existing
sensor, the connection manager creates a new sensor with `status='pending'`.
The user can then approve or dismiss it via the UI or CLI.

## API Endpoints

### Brokers

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/mqtt/brokers` | `view_mqtt` | List all brokers |
| POST | `/api/mqtt/brokers` | `manage_mqtt` | Create a broker |
| GET | `/api/mqtt/brokers/:id` | `view_mqtt` | Get broker by ID |
| PUT | `/api/mqtt/brokers/:id` | `manage_mqtt` | Update a broker |
| DELETE | `/api/mqtt/brokers/:id` | `manage_mqtt` | Delete a broker |

### Subscriptions

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/mqtt/subscriptions` | `view_mqtt` | List subscriptions (optional `?broker_id=N`) |
| POST | `/api/mqtt/subscriptions` | `manage_mqtt` | Create a subscription |
| GET | `/api/mqtt/subscriptions/:id` | `view_mqtt` | Get subscription by ID |
| PUT | `/api/mqtt/subscriptions/:id` | `manage_mqtt` | Update a subscription |
| DELETE | `/api/mqtt/subscriptions/:id` | `manage_mqtt` | Delete a subscription |

### Sensor Status

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/sensors/status/:status` | `view_sensors` | List sensors by status |
| POST | `/api/sensors/approve/:id` | `manage_sensors` | Approve a pending sensor |
| POST | `/api/sensors/dismiss/:id` | `manage_sensors` | Dismiss a pending sensor |

### Permissions

Two permissions were added in migration 000008:

| Permission | Granted to | Purpose |
|------------|-----------|---------|
| `view_mqtt` | All roles | View brokers and subscriptions |
| `manage_mqtt` | Admin only | Create, update, delete brokers/subscriptions |

## Writing a New PushDriver

PushDrivers are stateless message parsers. They don't own MQTT connections —
the connection manager handles that. A PushDriver only needs to:

1. Parse an MQTT payload into readings
2. Identify which device sent the message (for auto-discovery)

### The PushDriver Interface

```go
type PushDriver interface {
    SensorDriver  // Type, DisplayName, Description, ConfigFields, SupportedMeasurementTypes, ValidateSensor

    // ParseMessage extracts readings from an MQTT message payload.
    ParseMessage(topic string, payload []byte) ([]types.Reading, error)

    // IdentifyDevice returns a suggested sensor name from an MQTT message,
    // used during auto-discovery of new devices.
    IdentifyDevice(topic string, payload []byte) (string, error)
}
```

### Step-by-Step

#### 1. Create the driver file

Create `sensor_hub/drivers/mqtt_your_ecosystem.go`:

```go
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
    Register(&YourEcosystemDriver{})
}

type YourEcosystemDriver struct{}

var _ PushDriver = (*YourEcosystemDriver)(nil) // compile-time check

func (d *YourEcosystemDriver) Type() string        { return "mqtt-your-ecosystem" }
func (d *YourEcosystemDriver) DisplayName() string  { return "Your Ecosystem (MQTT)" }
func (d *YourEcosystemDriver) Description() string {
    return "Devices reporting via Your Ecosystem MQTT bridge"
}

// Push drivers typically have no per-sensor config fields.
func (d *YourEcosystemDriver) ConfigFields() []ConfigFieldSpec { return nil }

func (d *YourEcosystemDriver) SupportedMeasurementTypes() []types.MeasurementType {
    return []types.MeasurementType{
        {Name: "temperature", DisplayName: "Temperature", Unit: "°C", Category: "numeric"},
    }
}

// Push-based sensors have nothing to validate locally.
func (d *YourEcosystemDriver) ValidateSensor(_ context.Context, _ types.Sensor) error {
    return nil
}

func (d *YourEcosystemDriver) IdentifyDevice(topic string, _ []byte) (string, error) {
    // Extract device name from topic structure.
    // e.g., "your-ecosystem/devices/living-room" → "living-room"
    parts := strings.Split(topic, "/")
    if len(parts) < 3 {
        return "", fmt.Errorf("unexpected topic structure: %s", topic)
    }
    return parts[len(parts)-1], nil
}

func (d *YourEcosystemDriver) ParseMessage(topic string, payload []byte) ([]types.Reading, error) {
    var data map[string]interface{}
    if err := json.Unmarshal(payload, &data); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    now := time.Now().UTC().Format("2006-01-02 15:04:05")

    tempVal, ok := data["temperature"].(float64)
    if !ok {
        return nil, fmt.Errorf("missing temperature field")
    }

    return []types.Reading{{
        MeasurementType: "temperature",
        NumericValue:    &tempVal,
        Unit:            "°C",
        Time:            now,
    }}, nil
}
```

#### 2. Write tests

Create `sensor_hub/drivers/mqtt_your_ecosystem_test.go` with tests covering:

- Metadata (Type, DisplayName, SupportedMeasurementTypes)
- `IdentifyDevice` with valid and invalid topics
- `ParseMessage` with realistic payloads
- Edge cases (missing fields, invalid JSON, system topics)
- Compile-time interface check: `var _ PushDriver = (*YourEcosystemDriver)(nil)`

#### 3. Add measurement types (if needed)

If your ecosystem introduces measurement types not already in the database,
create a new migration in `db/migrations/`.

#### 4. Create a subscription

After deploying the driver, create a subscription that routes messages to it:

```bash
sensor-hub mqtt subscriptions create \
  --broker-id 1 \
  --topic "your-ecosystem/#" \
  --driver mqtt-your-ecosystem
```

That's all that's needed — no changes to the service layer, API, or UI.

### Key differences from PullDriver

| Aspect | PullDriver | PushDriver |
|--------|------------|------------|
| Data flow | Service calls `CollectReadings()` on timer | Connection manager calls `ParseMessage()` on message arrival |
| Connection | Driver manages its own connections | Connection manager manages shared MQTT clients |
| Config fields | Per-sensor config (URLs, credentials) | Typically none (connection info is on the broker record) |
| Validation | `ValidateSensor` may test connectivity | `ValidateSensor` is usually a no-op |
| `SensorName` on readings | Set by driver (from `sensor.Name`) | Set by connection manager (from `IdentifyDevice`) |

## Existing PushDriver: Zigbee2MQTT

| Property | Value |
|----------|-------|
| File | `drivers/zigbee2mqtt.go` |
| Type | `mqtt-zigbee2mqtt` |
| Topic pattern | `zigbee2mqtt/#` |
| Measurement types | 23 types (temperature, humidity, pressure, battery, voltage, illuminance, power, energy, current, co2, voc, contact, occupancy, and more) |

The Zigbee2MQTT driver uses a field mapping registry (`knownFields`) that maps
Zigbee2MQTT's normalised JSON field names to measurement type definitions.
Unknown fields are silently ignored, so it handles arbitrary Zigbee hardware
without code changes.

Topic parsing extracts the device friendly name from the last segment(s) of the
topic. System topics (`bridge/*`), command topics (`*/set`, `*/get`), and
availability topics are filtered out.

## Connection Manager

The connection manager (`mqtt/connection_manager.go`) is the runtime bridge
between MQTT brokers and the rest of the system. It:

- Maintains one Paho MQTT client per enabled broker
- Subscribes to all enabled subscriptions for each broker
- Re-subscribes automatically on reconnect
- Routes incoming messages: topic → subscription → PushDriver → SensorService
- Auto-discovers unknown devices as pending sensors

### Startup

Called in `serve.go` after sensor discovery and before periodic collection:

```go
connManager := mqttBrokerPkg.NewConnectionManager(sensorService, mqttSubRepo, mqttBrokerRepo, logger)
if err := connManager.Start(ctx); err != nil {
    logger.Error("failed to start MQTT connection manager", "error", err)
}
defer connManager.Stop()
```

### Message Routing

When a message arrives:

1. Find all subscriptions whose `topic_pattern` matches (MQTT wildcard matching)
2. For each matching subscription, look up the `PushDriver` via `drivers.Get(sub.DriverType)`
3. Call `driver.IdentifyDevice(topic, payload)` to get the device name
4. Look up the sensor by name; if not found, auto-create with `status='pending'`
5. If sensor is dismissed, skip processing
6. Call `driver.ParseMessage(topic, payload)` to get readings
7. Call `sensorService.ServiceProcessPushReadings(ctx, sensorName, readings)`

## Embedded Broker

The embedded MQTT broker (`mqtt/broker.go`) wraps mochi-mqtt v2. It runs
inside the Sensor Hub process and listens on a configurable TCP port.

It is optional — you can use only external brokers. Enable it with:

```properties
mqtt.broker.enabled=true
mqtt.broker.port=1883
```

The embedded broker starts before the database is opened and stops during
graceful shutdown. It supports standard MQTT 3.1.1 and 5.0 clients.

## Database Schema

Migration `000008_mqtt_ingest` adds:

- `mqtt_brokers` table
- `mqtt_subscriptions` table (with FK to mqtt_brokers)
- `sensors.status` column (default `'active'`)
- `view_mqtt` and `manage_mqtt` permissions
