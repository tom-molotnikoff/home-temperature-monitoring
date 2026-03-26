#!/bin/bash
rm -f openapi.yaml
cp ../temperature_sensor/openapi.yaml openapi.yaml
docker compose down --remove-orphans
docker compose up --build