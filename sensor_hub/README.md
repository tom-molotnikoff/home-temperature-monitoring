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

There is a docker_tests folder which has an almost exactly similar setup, except it uses mock-sensors to provide data without reliance on actual infrastructure, and the sensor-hub container uses Air and Delve for Live Hotswapping of changes in the Go Application.

```sh
cd docker_tests
docker compose up --build
```

## Debug setup

Since the project uses Air and Delve for the debugging of the Go application, you can attach a debugger to dive into the call stack and variables.

Create a launch.json in the .vscode directory with the following contents:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Connect to container",
      "type": "go",
      "debugAdapter": "dlv-dap",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "localhost",
      "trace": "verbose",
      "substitutePath": [
        {
          "from": "/Users/tommolotnikoff/Documents/personal/git/home-temperature-monitoring/sensor_hub",
          "to": "/app"
        }
      ]
    }
  ]
}
```

Replace the path to the "sensor_hub" folder with the appropriate path on the system you are on. Run the debugger and off u go.

Atm, exiting the debugger is closing the container for sensor_hub - to be fixed
