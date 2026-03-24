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

## Software installation

1. Install Python prerequisites:

```bash
sudo apt install python3-pip python3.11-venv
```

2. Clone the repository and navigate to the sensor directory:

```bash
git clone <repo-url>
cd home-temperature-monitoring/temperature_sensor
```

3. Create a Python virtual environment and install dependencies:

```bash
python3 -m venv ./venv
venv/bin/pip3 install -r requirements.txt
```

4. Create a `.env` file in the sensor directory:

```bash
FLASK_APP=/home/<user>/home-temperature-monitoring/temperature_sensor/sensor_api.py
```

Replace the path with the actual location of `sensor_api.py` on your device.

## Running as a systemd service

The repository includes a systemd unit file that runs the sensor API as a background service.

1. Edit `temp_sensor_api.service` to update the paths and user/group to match your system:

```ini
[Unit]
Description=API for connected temperature sensor
After=network.target auditd.service

[Service]
EnvironmentFile=/home/<user>/home-temperature-monitoring/temperature_sensor/.env
User=<user>
Group=<user>
ExecStart=/home/<user>/home-temperature-monitoring/temperature_sensor/venv/bin/python3 -m flask run --host=0.0.0.0
Restart=on-failure
RestartPreventExitStatus=255

[Install]
WantedBy=multi-user.target
```

2. Copy the unit file to systemd and enable the service:

```bash
sudo cp temp_sensor_api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now temp_sensor_api
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

To update the sensor software after a new release:

```bash
cd home-temperature-monitoring/temperature_sensor
git pull
venv/bin/pip3 install -r requirements.txt
sudo systemctl restart temp_sensor_api
```

## Registering the sensor in Sensor Hub

After deploying a sensor, register it in Sensor Hub so that it is polled for readings. See [Managing Sensors](managing-sensors) for details on registering sensors through the UI or API.
