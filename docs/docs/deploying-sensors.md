---
id: deploying-sensors
title: Deploying Sensors
sidebar_position: 5
---

# Deploying Sensors

Each temperature sensor is a lightweight HTTP service that runs on a Raspberry Pi and reads a DS18B20 sensor over the 1-wire protocol. Sensor Hub polls these endpoints to collect readings.

## How sensors work

A sensor exposes a single HTTP endpoint that returns the current temperature reading. Sensor Hub makes a `GET` request to this endpoint at a configurable interval (default: every 5 minutes). The sensor does not push data; all communication is initiated by Sensor Hub.

Sensors are expected to be on a trusted local network. The sensor API does not implement authentication.

## Hardware setup

1. Connect a DS18B20 temperature sensor to your Raspberry Pi using the 1-wire protocol.
2. Enable the 1-wire interface on the Raspberry Pi. You can do this through `raspi-config` or by adding the following line to `/boot/config.txt`:

```
dtoverlay=w1-gpio
```

3. Reboot the Raspberry Pi for the change to take effect.

## Prerequisites

Before installing the sensor package, ensure the following packages are installed
on the Raspberry Pi:

```bash
sudo apt update
sudo apt install -y python3 python3-venv python3-pip
```

These are required for the sensor's Python virtual environment. The `dpkg` installer
will **not** install them automatically — if they are missing, the package will be
left in an unconfigured state and must be fixed with `sudo apt install -f`.

## Installation

Download the latest `temperature-sensor` `.deb` package from the [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases) page and install it:

```bash
wget https://github.com/tom-molotnikoff/home-temperature-monitoring/releases/download/vVERSION/temperature-sensor_VERSION_all.deb
sudo apt install ./temperature-sensor_VERSION_all.deb
```

Replace `VERSION` with the version you are installing (e.g. `1.0.0`).

The package automatically:
- Installs the sensor API to `/usr/lib/temperature-sensor/`
- Creates a Python virtual environment and installs all dependencies
- Creates a dedicated `temperature-sensor` system user with `gpio` group access for hardware reading
- Installs and enables a systemd service (`temperature-sensor.service`)

:::note
The Pi requires an internet connection during installation so that `pip` can download the Python dependencies into the virtual environment. This may take several minutes on lower-powered Pis (e.g. Pi Zero) as some C extensions are compiled from source.
:::

### Verifying the GPG signature

Each release includes a detached GPG signature (`.deb.sig`). To verify:

```bash
# Import the public key (one-time)
wget https://github.com/tom-molotnikoff/home-temperature-monitoring/releases/download/vVERSION/sensor-hub-gpg-public.key
gpg --import sensor-hub-gpg-public.key

# Verify the package
gpg --verify temperature-sensor_VERSION_all.deb.sig temperature-sensor_VERSION_all.deb
```

## Configuration

Edit `/etc/temperature-sensor/environment` to configure the sensor:

```bash
# Change the Flask port (default: 5000)
# FLASK_RUN_PORT=5000

# Enable OpenTelemetry distributed tracing (optional)
# OTEL_EXPORTER_OTLP_ENDPOINT=http://your-collector:4317
# OTEL_SERVICE_NAME=temperature-sensor-living-room
```

The `OTEL_SERVICE_NAME` is a good place to give each sensor a human-readable name (e.g. `temperature-sensor-kitchen`). When OpenTelemetry is configured, the sensor participates in distributed traces — you can see the full request lifecycle from Sensor Hub through to the sensor in tools like Grafana Tempo.

## Starting the service

```bash
sudo systemctl start temperature-sensor
sudo systemctl status temperature-sensor
```

The sensor API starts on port 5000 by default and listens on all network interfaces.

## Sensor API response

The sensor returns a JSON object with the current timestamp and temperature in Celsius:

```json
{
  "time": "2026-01-15 14:30:00",
  "temperature": 21.56
}
```

If the sensor cannot be read, the API returns a 500 status code with an error message.

## Verifying a sensor

Test that the sensor is responding:

```bash
curl http://<sensor-ip>:5000/temperature
```

## Updating a sensor

Download and install the new `.deb` package. The upgrade is handled automatically — dependencies are reinstalled and the service is restarted:

```bash
wget https://github.com/tom-molotnikoff/home-temperature-monitoring/releases/download/vNEW_VERSION/temperature-sensor_NEW_VERSION_all.deb
sudo apt install ./temperature-sensor_NEW_VERSION_all.deb
```

Your configuration in `/etc/temperature-sensor/environment` is preserved across upgrades.

## Uninstalling

```bash
# Remove the package (keeps configuration files)
sudo apt remove temperature-sensor

# Purge the package (removes everything including config and virtualenv)
sudo apt purge temperature-sensor
```

## Registering the sensor in Sensor Hub

After deploying a sensor, register it in Sensor Hub so that it is polled for readings. See [Managing Sensors](managing-sensors) for details on registering sensors through the UI or API.
