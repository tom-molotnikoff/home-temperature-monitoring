---
id: prerequisites
title: Prerequisites
sidebar_position: 2
---

# Prerequisites

Before installing Sensor Hub, ensure your environment meets the following requirements.

## Host system

- Docker Engine 20.10 or later
- Docker Compose v2
- Sufficient disk space for SQLite database storage (depends on the number of sensors and data retention settings)

## Network

- Port 8080 available for the backend API
- Port 3000 available for the web UI (HTTP)
- Port 3443 available for the web UI (HTTPS, production only)


## TLS certificates

For deployments exposed to the internet, you need TLS certificates in PEM format:

- A certificate file (e.g., `home.sensor-hub.pem`)
- A private key file (e.g., `home.sensor-hub-key.pem`)
- A CA certificate if using self-signed certificates (e.g., for Nginx to verify the backend)

For local development, [mkcert](https://github.com/FiloSottile/mkcert) can generate locally-trusted certificates.

## Temperature sensors

To deploy temperature sensors, each sensor node requires:

- A Raspberry Pi (or similar Linux single-board computer) with network access to the Sensor Hub host
- A DS18B20 temperature sensor connected via the 1-wire protocol
- The 1-wire interface enabled on the device (via `raspi-config` or by adding `dtoverlay=w1-gpio` to `/boot/config.txt`)
- Python 3.11 or later with pip and venv

## Email notifications (optional)

To send alert notifications via email, you need:

- A Google Cloud project with the Gmail API enabled
- An OAuth 2.0 credential of type "Desktop application"
- The `credentials.json` file downloaded from the Google Cloud Console

The OAuth token is obtained during setup using a provided authorization tool or through the web UI after deployment.
