---
id: sensors
title: Connecting Sensors
sidebar_position: 5
---

# Connecting Sensors

Sensor Hub collects data from physical sensors using a **driver** model. Each sensor driver knows how to communicate with a particular type of hardware or protocol. The system currently ships with two drivers, covering HTTP-polled sensors and Zigbee devices via MQTT.

## How sensors connect

Sensors connect to Sensor Hub in one of two ways, depending on the driver:

| Model | How it works | Example |
|-------|-------------|---------|
| **Pull** (HTTP polling) | Sensor Hub makes an HTTP request to the sensor at a regular interval and reads the response | [HTTP Temperature Sensor](http-temperature) |
| **Push** (MQTT) | The sensor publishes messages to an MQTT broker. Sensor Hub subscribes and processes messages as they arrive | [Zigbee devices via Zigbee2MQTT](zigbee) |

Both models feed into the same data pipeline — readings are stored, alerts are evaluated, and real-time updates are broadcast to connected UI clients via WebSocket, regardless of how the data was collected.

## Supported drivers

| Driver | Type | Protocol | Measurements | Details |
|--------|------|----------|-------------|---------|
| `sensor-hub-http-temperature` | Pull | HTTP | Temperature (°C) | [HTTP Temperature Sensor](http-temperature) |
| `mqtt-zigbee2mqtt` | Push | MQTT via Zigbee2MQTT | 25+ types — temperature, humidity, power, energy, contact, occupancy, and more | [Zigbee (via Zigbee2MQTT)](zigbee) |

You can see the available drivers and their configuration schemas at any time via:

- The web UI **Add Sensor** form (config fields appear dynamically when you select a driver)
- The REST API: `GET /api/drivers`
- The CLI: `sensor-hub drivers list`

## Auto-discovery (push sensors)

MQTT-based sensors are **auto-discovered**. When a new device publishes a message that matches a configured MQTT subscription, Sensor Hub automatically creates a sensor entry with a **Pending** status. You then approve or dismiss the sensor through the UI or CLI. This means you don't need to manually register each Zigbee device — just pair it with Zigbee2MQTT and it appears in Sensor Hub.

Pull-based sensors (HTTP) must be registered manually through the UI, API, or CLI.

## How-to guides

Looking for step-by-step instructions?

- [How to connect an HTTP temperature sensor](../how-to/connect-http-sensor) — register a pull-based sensor in Sensor Hub
- [How to connect a Zigbee device to Sensor Hub](../how-to/connect-zigbee-device) — set up Zigbee2MQTT and connect your first Zigbee device
- [How to monitor energy usage with a Zigbee smart plug](../how-to/monitor-energy-usage) — build a power monitoring dashboard

## After connecting

Once sensors are connected and reporting data, see [Managing Sensors](managing-sensors-ref) for information on data collection, health monitoring, data retention, and permissions.
