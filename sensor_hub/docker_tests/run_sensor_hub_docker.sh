#!/bin/bash
rm openapi.yaml
cp ../temperature_sensor/openapi.yaml openapi.yaml
docker compose down --volumes --remove-orphans
docker compose up --build