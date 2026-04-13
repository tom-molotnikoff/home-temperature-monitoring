---
id: connect-http-sensor
title: How to connect an HTTP temperature sensor
sidebar_position: 3
---

# How to connect an HTTP temperature sensor

This guide walks you through registering an HTTP temperature sensor in Sensor Hub so that readings are collected automatically.

## Before you start

You will need:

- A running Sensor Hub instance (see [Installation](../installation))
- A temperature sensor accessible on your network that exposes an HTTP endpoint returning JSON in this format:

```json
{
  "time": "2026-01-15 14:30:00",
  "temperature": 21.56
}
```

If you are using the Sensor Hub temperature sensor package on a Raspberry Pi, see the [HTTP Temperature Sensor](../sensors/http-temperature#sensor-hub-temperature-sensor-package) reference for installation instructions.

## Step 1 — Verify the sensor is responding

From the machine running Sensor Hub, test that the sensor is reachable:

```bash
curl http://<sensor-ip>:5000/temperature
```

You should see a JSON response with `temperature` and `time` fields. If you get a connection error, check that:

- The sensor service is running on the remote machine
- The sensor's IP and port are correct
- No firewall is blocking the connection between Sensor Hub and the sensor

:::tip
If Sensor Hub is running in Docker, the sensor must be reachable from **inside** the Docker network. You may need to use the host's IP address rather than `localhost`.
:::

## Step 2 — Register the sensor

### Using the web UI

1. Navigate to **Sensors** in the Sensor Hub web UI
2. Click **Add Sensor**
3. Fill in the form:
   - **Name**: A friendly name for the sensor (e.g. "Living Room", "Upstairs Bedroom")
   - **Driver**: Select **Sensor Hub HTTP Temperature**
   - **URL**: The base URL of the sensor (e.g. `http://192.168.1.50:5000`) — do not include `/temperature`, the driver appends it automatically
4. Click **Save**

### Using the CLI

```bash
sensor-hub sensors add \
  --name "Living Room" \
  --driver sensor-hub-http-temperature \
  --config url=http://192.168.1.50:5000
```

### Using the API

```bash
curl -X POST http://localhost:8080/api/sensors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Living Room",
    "sensor_driver": "sensor-hub-http-temperature",
    "config": {
      "url": "http://192.168.1.50:5000"
    }
  }'
```

When registering via the API or CLI, Sensor Hub validates the sensor by attempting to fetch a reading. If the sensor is not reachable, the registration fails with an error.

## Step 3 — Verify readings are appearing

After registering the sensor:

1. Navigate to **Sensors** in the web UI
2. Click on your new sensor
3. Within one polling interval (default: 5 minutes), you should see the first temperature reading

You can also trigger an immediate reading:

**Web UI:** Click the **Refresh** button on the sensor page.

**CLI:**

```bash
sensor-hub sensors read --name "Living Room"
```

## Troubleshooting

**Sensor shows "Bad" health status:**

- Verify the sensor is still responding: `curl http://<sensor-ip>:5000/temperature`
- Check the Sensor Hub logs for error details: `journalctl -u sensor-hub -f`
- If the sensor URL changed, update it via the sensor's edit page or the API

**Sensor reachable from host but not from Sensor Hub (Docker):**

- Use the host machine's LAN IP (e.g. `http://192.168.1.10:5000`) instead of `localhost`
- Alternatively, use `http://host.docker.internal:5000` if your Docker setup supports it

**Readings appear but are stale:**

- Check the polling interval — the default is 5 minutes. You can lower it via the `sensor.collection.interval` property (see [Configuration Settings](../configuration))
