---
id: sensors
title: Connecting Sensors
sidebar_position: 5
---

# Connecting Sensors

Sensor Hub collects data from physical sensors using a **driver** model. Each sensor driver knows how to communicate with a particular type of hardware or protocol. 

Notably, Sensor Hub supports Zigbee devices via MQTT through the popular Zigbee2MQTT project. This means you can connect a wide range of Zigbee sensors and devices without needing a separate Zigbee hub — just set up Zigbee2MQTT and Sensor Hub will auto-discover your devices as they publish MQTT messages.

## How sensors connect

Sensors connect to Sensor Hub in one of two ways, depending on the driver:

| Model                   | How it works                                                                                                 | Example                                     |
|-------------------------|--------------------------------------------------------------------------------------------------------------|---------------------------------------------|
| **Pull** (HTTP polling) | Sensor Hub makes an HTTP request to the sensor at a regular interval and reads the response                  | [HTTP Temperature Sensor](http-temperature) |
| **Push** (MQTT)         | The sensor publishes messages to an MQTT broker. Sensor Hub subscribes and processes messages as they arrive | [Zigbee devices via Zigbee2MQTT](zigbee)    |

Both models feed into the same data pipeline — readings are stored, alerts are evaluated, and real-time updates are broadcast to connected UI clients via WebSocket, regardless of how the data was collected.

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

If your device can be switched or otherwise commanded, see [Device Control](device-control) for controllable sensor capabilities, dashboard setup, permissions, and troubleshooting.
