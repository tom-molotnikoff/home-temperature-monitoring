#!/bin/bash
cd docker
docker compose down --volumes --remove-orphans
docker compose up --build -d