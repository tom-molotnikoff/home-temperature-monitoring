---
id: zigbee
title: Zigbee (via Zigbee2MQTT)
sidebar_position: 3
---

# Zigbee (via Zigbee2MQTT)

The `mqtt-zigbee2mqtt` driver connects Zigbee devices to Sensor Hub through [Zigbee2MQTT](https://www.zigbee2mqtt.io/), an open-source Zigbee-to-MQTT bridge. It is a **push-based** driver — devices publish MQTT messages as readings change, and Sensor Hub processes them in real time.

This driver supports a wide range of Zigbee hardware without any code changes. Unknown device fields are silently ignored, so new Zigbee devices generally work out of the box as long as Zigbee2MQTT supports them.

## Driver details

| Property | Value |
|----------|-------|
| Driver type | `mqtt-zigbee2mqtt` |
| Protocol | MQTT (via Zigbee2MQTT bridge) |
| Collection model | Push (messages arrive via MQTT subscription) |
| Config fields | None (push drivers have no per-sensor configuration) |
| Typical MQTT topic | `zigbee2mqtt/#` |

## Supported measurement types

The driver extracts all recognised fields from the Zigbee2MQTT JSON payload. The measurements available depend on your device.

### Numeric measurements

| Measurement | Unit | Typical devices |
|------------|------|-----------------|
| Temperature | °C | Temperature sensors, multi-sensors, thermostats |
| Humidity | % | Humidity sensors, multi-sensors |
| Pressure | hPa | Weather stations, air quality sensors |
| Battery | % | All battery-powered devices |
| Voltage | V | Battery-powered devices, smart plugs |
| Link quality | lqi | All Zigbee devices |
| Illuminance | lx | Motion sensors, light sensors |
| Power | W | Smart plugs, energy monitors |
| Energy | kWh | Smart plugs, energy monitors |
| Energy today | kWh | Smart plugs with daily tracking |
| Energy yesterday | kWh | Smart plugs with daily tracking |
| Energy month | kWh | Smart plugs with monthly tracking |
| Current | A | Smart plugs, energy monitors |
| CO₂ | ppm | Air quality sensors |
| VOC | ppb | Air quality sensors |
| Formaldehyde | mg/m³ | Air quality sensors |
| PM2.5 | µg/m³ | Air quality sensors |
| Soil moisture | % | Soil sensors |

### Binary measurements

| Measurement | States | Typical devices |
|------------|--------|-----------------|
| Contact | open / closed | Door/window contact sensors |
| Occupancy | true / false | Motion sensors, presence sensors |
| Water leak | true / false | Water leak sensors |
| Smoke | true / false | Smoke detectors |
| Carbon monoxide | true / false | CO detectors |
| Tamper | true / false | Security sensors |
| Battery low | true / false | Battery-powered devices |
| Vibration | true / false | Vibration sensors |
| State | on / off | Switches, relays |

## How auto-discovery works

When a Zigbee device publishes a message to the MQTT broker, Sensor Hub processes it through the following flow:

1. The message arrives on a topic matching the configured subscription pattern (typically `zigbee2mqtt/#`)
2. The driver extracts the **device name** from the MQTT topic — this is the friendly name you assigned in Zigbee2MQTT
3. Sensor Hub looks up the device by its external ID (or name as a fallback)
4. If the device is **unknown**, a new sensor is created with **Pending** status
5. Once you **approve** the sensor (via the UI or CLI), readings are collected and stored

This means you never need to manually register Zigbee devices. Pair a device with Zigbee2MQTT, give it a friendly name, and it appears in Sensor Hub automatically.

### Sensor statuses

| Status | Meaning |
|--------|---------|
| **Pending** | Auto-discovered, awaiting your approval |
| **Active** | Approved — readings are collected and stored |
| **Dismissed** | You chose to ignore this device (can be restored later) |

### Approving and dismissing sensors

**Web UI:** Navigate to **Sensors**, find the pending sensor, and click **Approve** or **Dismiss**.

**CLI:**

```bash
# List pending sensors
sensor-hub sensors list --status pending

# Approve a sensor
sensor-hub sensors approve --id <sensor-id>

# Dismiss a sensor
sensor-hub sensors dismiss --id <sensor-id>
```

## MQTT subscription configuration

For the driver to receive messages, you need an MQTT subscription that routes Zigbee2MQTT traffic to it.

### Embedded broker

Sensor Hub includes an embedded MQTT broker (enabled by default on port 1883). Point Zigbee2MQTT at this broker and no external MQTT infrastructure is needed.

### Creating the subscription

**Web UI:**

1. Navigate to **Settings → MQTT Subscriptions**
2. Click **Add Subscription** and configure:
   - **Broker**: Select the embedded broker
   - **Topic pattern**: `zigbee2mqtt/#`
   - **Driver type**: `mqtt-zigbee2mqtt`
3. Save

**CLI:**

```bash
sensor-hub mqtt subscriptions add \
  --broker-id 1 \
  --topic "zigbee2mqtt/#" \
  --driver mqtt-zigbee2mqtt
```

## Device naming

The sensor name in Sensor Hub comes from the **friendly name** you assign to the device in Zigbee2MQTT. To rename a device:

1. Open the Zigbee2MQTT web UI
2. Click on the device and change its friendly name
3. The new name is reflected in Sensor Hub on the next message

Sensor Hub uses a stable `external_id` to track devices, so renaming a sensor in the Sensor Hub UI does not break the link to the MQTT device.

## How-to guide

For a complete step-by-step walkthrough — including installing Zigbee2MQTT, pairing devices, and configuring the subscription — see [How to connect a Zigbee device to Sensor Hub](../how-to/connect-zigbee-device).
