---
id: connect-zigbee-device
title: How to connect a Zigbee device to Sensor Hub
sidebar_position: 1
---

# How to connect a Zigbee device to Sensor Hub

This guide walks you through connecting a Zigbee device — such as a smart plug, temperature sensor, or contact sensor — to Sensor Hub using Zigbee2MQTT.

## Before you start

You will need:

- A running Sensor Hub instance with MQTT enabled (`mqtt.broker.enabled=true`, the default)
- A Zigbee USB coordinator (e.g. SONOFF Zigbee 3.0 Dongle Plus ZBDongle-E, ConBee II, or similar)
- A Zigbee-compatible device (e.g. SONOFF S40 Lite smart plug, Aqara temperature sensor, SONOFF SNZB-02 sensor)
- A machine to run Zigbee2MQTT (the same machine as Sensor Hub, or any machine on the same network)

## Step 1 — Install Zigbee2MQTT

Zigbee2MQTT is an open-source bridge that translates Zigbee radio traffic into MQTT messages. Install it on a machine with the USB coordinator plugged in.

**Find your coordinator device path:**

Plug in the USB coordinator and run:

```bash
ls -la /dev/serial/by-id/
```

You will see output like:

```
usb-ITEAD_SONOFF_Zigbee_3.0_USB_Dongle_Plus_V2_... -> ../../ttyACM0
```

or:

```
usb-Silicon_Labs_Sonoff_Zigbee_3.0_USB_Dongle_Plus_... -> ../../ttyUSB0
```

Note the full `/dev/serial/by-id/...` path — using this instead of `/dev/ttyUSB0` ensures the device is identified correctly even if USB port numbering changes after a reboot.

**Using Docker (recommended):**

```yaml
# docker-compose.yml
services:
  zigbee2mqtt:
    image: koenkk/zigbee2mqtt
    restart: unless-stopped
    volumes:
      - ./zigbee2mqtt-data:/app/data
      - /run/udev:/run/udev:ro
    ports:
      - 8081:8080
    devices:
      - /dev/serial/by-id/YOUR_COORDINATOR_ID:/dev/ttyACM0
    environment:
      - TZ=Europe/London
```

Replace `YOUR_COORDINATOR_ID` with the full path from the `ls` command above (e.g. `usb-ITEAD_SONOFF_Zigbee_3.0_USB_Dongle_Plus_V2_...`).

**Or follow the [official Zigbee2MQTT installation guide](https://www.zigbee2mqtt.io/guide/installation/).**

## Step 2 — Configure Zigbee2MQTT to publish to Sensor Hub

Edit the Zigbee2MQTT configuration file (`zigbee2mqtt-data/configuration.yaml`) to point at Sensor Hub's MQTT broker:

```yaml
mqtt:
  base_topic: zigbee2mqtt
  server: mqtt://SENSOR_HUB_HOST:1883

serial:
  port: /dev/ttyACM0

frontend:
  enabled: true
  port: 8080

advanced:
  log_level: info
```

Replace `SENSOR_HUB_HOST` with the IP address or hostname of your Sensor Hub machine. The `serial.port` should match the right-hand side of the Docker `devices` mapping (i.e. `/dev/ttyACM0`). If you are running Zigbee2MQTT without Docker, use the full `/dev/serial/by-id/...` path instead.

If you are running Zigbee2MQTT on the same machine as Sensor Hub, use `mqtt://localhost:1883`.

Restart Zigbee2MQTT after editing the configuration.

## Step 3 — Pair your Zigbee device

1. Open the Zigbee2MQTT web UI (default: `http://ZIGBEE2MQTT_HOST:8081`)
2. Click **Permit join** to enable pairing mode
3. Put your device into pairing mode:
   - **SONOFF S40 Lite**: Hold the button for 5 seconds until the LED flashes
   - **Aqara sensors**: Hold the reset button for 5 seconds
   - Check your device manual for specific instructions
4. The device should appear in the Zigbee2MQTT dashboard within 30 seconds
5. Give it a friendly name (e.g. `office-plug`, `living-room-sensor`) — this becomes the sensor name in Sensor Hub

## Step 4 — Add an MQTT subscription in Sensor Hub

In the Sensor Hub web UI:

1. Navigate to **Settings → MQTT Brokers**
2. The embedded broker should already be listed. If not, check that `mqtt.broker.enabled=true` in your configuration
3. Navigate to **Settings → MQTT Subscriptions**
4. Click **Add Subscription** and fill in:
   - **Broker**: Select the embedded broker
   - **Topic pattern**: `zigbee2mqtt/#`
   - **Driver type**: `mqtt-zigbee2mqtt`
5. Save the subscription

**Using the CLI:**

```bash
sensor-hub mqtt subscriptions add \
  --broker-id 1 \
  --topic "zigbee2mqtt/#" \
  --driver mqtt-zigbee2mqtt
```

## Step 5 — Verify the device appears

Once a Zigbee device publishes its first message, Sensor Hub auto-discovers it and creates a sensor entry with a **Pending** status.

1. Navigate to **Sensors** in the web UI
2. You should see the new device listed with its Zigbee2MQTT friendly name
3. Click **Approve** to activate the sensor
4. Readings will start appearing immediately

**Using the CLI:**

```bash
# List pending sensors
sensor-hub sensors list --status pending

# Approve a sensor
sensor-hub sensors approve --id <sensor-id>
```

## What gets collected

Sensor Hub automatically extracts all known measurement types from the Zigbee2MQTT JSON payload. The measurements available depend on your device:

| Device type | Typical measurements |
|------------|---------------------|
| Smart plug (e.g. SONOFF S40) | Power (W), Current (A), Voltage (mV), Energy (kWh) |
| Temperature/humidity sensor | Temperature (°C), Humidity (%), Battery (%) |
| Contact sensor | Contact (open/closed), Battery (%) |
| Motion sensor | Occupancy (true/false), Battery (%), Illuminance (lx) |
| Air quality sensor | CO₂ (ppm), VOC (ppb), Temperature (°C), Humidity (%) |

All devices also report **Link Quality** (lqi), which indicates Zigbee signal strength.

## Troubleshooting

**Device does not appear in Sensor Hub:**

- Check that Zigbee2MQTT is connected to the MQTT broker — the Zigbee2MQTT log should show `Connected to MQTT server`
- Verify the subscription topic matches: the default Zigbee2MQTT base topic is `zigbee2mqtt`, so `zigbee2mqtt/#` should catch everything
- Check the MQTT Broker Stats page in Sensor Hub to confirm messages are being received

**Device appears but shows no readings:**

- Ensure the sensor status is **Active** (approve it if it is still Pending)
- Some devices only report when a value changes — press the button or trigger the sensor to force an update

**Wrong device name:**

- Rename the device in the Zigbee2MQTT web UI. The new name will be reflected in Sensor Hub on the next message
