# Home temperature monitoring

## Overview

Preface: this is a pretty pointless project, you can already tell when your home is too hot or too cold, you don't need this over-engineered project to tell you.

With that out the way, this repo contains everything I use to monitor the temperature in my house. There are two raspberry pis with temperature sensors (DS18B20) attached to them. Those Pis run a small python application with a single endpoint API to collect the readings as JSON. There is a third, more powerful Pi, that runs "sensor hub". This service is responsible for aggregating the readings from the other Pis and persisting them into a MySQL database. This information is then served as a web page with a fancy graph.

## Sensors

The setup for the sensors is quite simple. Connecting a single sensor has been done lots before: <https://www.circuitbasics.com/raspberry-pi-ds18b20-temperature-sensor-tutorial/>.

The python API is defined in the /temperature_sensor portion of the repo. It is a very minimal flask API - genuinely a single endpoint. There is no persistence of the readings by the sensors themselves.

Ideally, this would probably not be an API, and instead be something like MQTT. I might change it to do that in future.

There is also legacy code in there for ingesting readings straight into google sheets. I did this initially as I didn't have anywhere to run a sensor-hub, this isn't used anymore, but it did work quite well for a long time.

## Sensor Hub

This is an application to aggregate and persist the readings from the sensors. It is containerised using docker compose, so it's very simple to reliably deploy on a Pi (I don't trust the SD cards not to fail). The MySQL data is held in a docker volume so it is available on the host. This can be backed up outside the Pi easily.

The whole of this project is retained inside my home network, the authentication of the database and sensors wasn't important to me, hence the rubbish mysql password setup and lack of API auth.

The backend application for Sensor Hub is written in Go, the frontend is written in typescript. The frontend is extremely rough, that part wasn't a passion project - it needs revisiting.

There is more information in the readme in the /sensor_hub folder. This is what the end result is though:

![image showing the dashboard of the sensor hub user interface](readme-assets/sensor-hub-dashboard.png "Sensor Hub Dashboard")
