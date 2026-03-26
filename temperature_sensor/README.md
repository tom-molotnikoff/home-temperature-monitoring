# Temperature Sensor

DS18B20 temperature sensor API for the Home Temperature Monitoring system. Runs as a Flask API on a Raspberry Pi, reading from a DS18B20 sensor over the 1-wire protocol.

## Installation (deb package)

Download the latest `.deb` from the [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases) page:

```bash
# Download (replace VERSION with the actual version)
wget https://github.com/tom-molotnikoff/home-temperature-monitoring/releases/download/vVERSION/temperature-sensor_VERSION_all.deb

# Install
sudo dpkg -i temperature-sensor_VERSION_all.deb
```

The package will:
- Install the sensor API to `/usr/lib/temperature-sensor/`
- Create a Python virtualenv and install dependencies
- Create a `temperature-sensor` system user (with `gpio` group access)
- Install a systemd service (`temperature-sensor.service`)

### Configuration

Edit `/etc/temperature-sensor/environment` to set environment variables:

```bash
# Optional: change the Flask port (default: 5000)
FLASK_RUN_PORT=5000

# Optional: enable OpenTelemetry distributed tracing
OTEL_EXPORTER_OTLP_ENDPOINT=http://your-collector:4317
OTEL_SERVICE_NAME=temperature-sensor-living-room
```

### Start the service

```bash
sudo systemctl start temperature-sensor
sudo systemctl status temperature-sensor
```

The sensor API will be available at `http://<pi-ip>:5000/temperature`.

### Updating

Download and install the new `.deb` — it will upgrade in place and restart the service.

### Uninstalling

```bash
sudo dpkg -r temperature-sensor    # Remove (keeps config)
sudo dpkg -P temperature-sensor    # Purge (removes everything)
```

## Prerequisites (hardware)

1. Raspberry Pi with PiOS
2. DS18B20 temperature sensor connected via 1-wire
3. 1-wire interface enabled: `sudo raspi-config` → Interface Options → 1-Wire → Enable

## Development setup

For local development without the deb package:

```bash
python3 -m venv ./venv
venv/bin/pip3 install -r requirements.txt
export FLASK_APP=sensor_api.py
venv/bin/flask run
```

## API

### `GET /temperature`

Returns the current temperature reading:

```json
{
  "temperature": 21.37,
  "time": "2026-03-26 16:00:00"
}
```
