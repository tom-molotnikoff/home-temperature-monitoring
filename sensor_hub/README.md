# Sensor Hub Application

This application is a web app that can collect readings from all sensors defined in an `openapi.yaml` file. These readings are stored in a MySQL database (separately running), and alerts are triggered when necessary.

## Configuration

- **Required:**

  - `database.properties` — Specify MySQL username, hostname, port, and password.
  - `application.properties` — Specify the location of the sensor's openapi.yaml file. This is used to identify the number of available sensors and their URLs.

- **Optional:**
  - `smtp.properties` — Configure SMTP settings to enable email alerts for threshold temperatures.
  - `application.properties` — Optionally set threshold temperatures for alerts.
  - `credentials.json` — If using SMTP for alerts, an appropriate credentials.json with a redirect uri of <http://localhost:8080> must be provided. This file should be put through the pre-authorisation script to generate a token.json before running sensor-hub. Sensor-hub cannot do the interactive authorisation process.

## Features

- API to collect readings from multiple sensors, or a specific sensor
- Stores data in a MySQL database
- Sends alerts when temperature thresholds are exceeded (if SMTP configured)

## Docker compose setup

In the docker folder there is a docker compose yaml to define a MySQL container and a Sensor Hub container. This can be spun up by running the run_sensor_hub_docker.sh script, or running the contained commands individually.
