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
│  Route message to PushDriver via subscription   │
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

## Configuration

### Embedded Broker

Sensor Hub can optionally run an embedded MQTT broker (mochi-mqtt) inside the
same process. This is useful for simple setups where you don't want to run
Mosquitto or another external broker.

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