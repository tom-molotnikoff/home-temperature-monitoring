---
id: prerequisites
title: Prerequisites
sidebar_position: 2
---

# Prerequisites

Before installing Sensor Hub, ensure your environment meets the following requirements.

## Host system

Supported operating systems:

- Fedora (latest stable)
- RHEL / CentOS Stream 9+
- Debian 12+
- Ubuntu 22.04+
- Raspberry Pi OS (arm64)

Additional requirements:

- nginx (for TLS termination)
- Sufficient disk space for the SQLite database (depends on the number of sensors and data retention settings)

## Network

- Port **443** — nginx (HTTPS, public-facing)
- Port **8080** — sensor-hub (localhost only, proxied by nginx)

## TLS certificates

You need TLS certificates in PEM format for nginx:

- A certificate file (e.g., `sensor-hub.pem`)
- A private key file (e.g., `sensor-hub-key.pem`)

For local development, [mkcert](https://github.com/FiloSottile/mkcert) can generate locally-trusted certificates. For production, use [Let's Encrypt](https://letsencrypt.org/) with certbot.

## Temperature sensors

Each sensor node requires:

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
