# temp-sensor-raspberry-pi

## Prerequisites

1. sudo apt install python3-pip python3.11-venv

## Installation

1. cd ~
2. git clone <git@github.com>:tom-molotnikoff/temp-sensor-raspberry-pi.git
3. cd temp-sensor-raspberry-pi
4. python3 -m venv ./venv
5. venv/bin/pip3 install -r requirements.txt

Create a .env file and fill with the needed values:\

```sh
FLASK_APP=/path/to/the/file/sensor_api.py
```

## Run an API to take readings

1. Edit the paths and copy the systemd unit file into /etc/systemd/system
2. systemctl daemon-reload
3. systemctl start temp_sensor_api

## Updating repo

1. pip freeze > requirements.txt
2. git commit... git push

## Updating on Pi

1. git pull
2. venv/bin/pip3 install -r requirements.txt
