# Sensor Drivers

Sensor Hub uses a **driver** model to support arbitrary sensor hardware and
protocols. Each driver is a single Go file that implements either the
`PullDriver` or `PushDriver` interface, self-registers at startup, and tells
the system how to communicate with a particular class of sensor.

## What is a Sensor Driver?

A sensor driver is the bridge between a physical (or virtual) sensor device and
the Sensor Hub data pipeline. It encapsulates:

- **Identity** — a unique type string, human-readable name, and description.
- **Measurement types** — which kinds of readings the sensor can produce
  (temperature, humidity, motion, etc.).
- **Configuration schema** — which config fields the driver needs (URLs, topics,
  credentials, etc.).
- **Data collection** — how to connect to the sensor and fetch readings.
- **Validation** — how to verify that a sensor's configuration is correct before
  the system starts collecting from it.

The rest of the application — storage, alerting, dashboards, WebSocket
broadcasts — is completely driver-agnostic. Once a driver returns a
reading, the generic pipeline takes over.

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

## The Driver Interfaces

Every driver implements the base `SensorDriver` interface plus one of the two
collection interfaces:

- **`PullDriver`** — for poll-based sensors where the service fetches data on a
  timer (e.g. HTTP temperature sensors).
- **`PushDriver`** — for event-driven MQTT ecosystems where messages arrive
  continuously (e.g. Zigbee2MQTT).