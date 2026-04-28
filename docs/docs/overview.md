---
id: overview
title: Overview
sidebar_position: 1
slug: /
---

# Overview

Sensor Hub is a self-hosted home monitoring system. It collects readings from sensors deployed on your network — temperature probes, Zigbee smart plugs, contact sensors, and more — stores historical data, and presents it through a responsive web interface with real-time updates.

## Capabilities

- Collect readings from HTTP-polled sensors and MQTT-based devices (Zigbee via Zigbee2MQTT)
- Auto-discover new devices as they appear on your MQTT network
- Store and visualise historical sensor data
- Support for measurement types including temperature, humidity, power, energy, contact, occupancy, and more
- Monitor sensor health and connectivity status
- Configure alert rules that trigger when readings exceed defined thresholds
- Receive notifications in-app and via email when alerts fire
- Build custom dashboards with configurable widgets
- Manage users, roles, and granular permissions
- Update system configuration at runtime through the UI or REST API
- Connect with AI assistants via CLI skills for natural language control of your monitoring system

## Architecture

Sensor Hub is distributed as a single Go binary (`sensor-hub`) packaged as RPM and DEB. The binary contains:

- A Go backend that exposes a REST API and WebSocket server on port 8080. All API routes live under the `/api` prefix. It handles sensor polling, data storage, alerting, authentication, and session management.
- A React/TypeScript single-page application embedded into the binary. The binary serves both the API and the UI, so no separate web server is needed.
- An embedded SQLite database at `/var/lib/sensor-hub/sensor_hub.db` that stores all persistent data including sensor readings, user accounts, sessions, alert rules, and notifications. Database schema changes are managed by golang-migrate embedded migrations that run automatically on startup.

Nginx sits in front of the binary as a TLS reverse proxy, forwarding HTTPS traffic on port 443 to the Go process on `127.0.0.1:8080`. Nginx is installed and configured separately — see the [nginx setup guide](nginx-setup).

## Get started

Review the [prerequisites](prerequisites) to prepare your environment, then follow the [installation guide](installation) to deploy Sensor Hub. Once running, see [Connecting Sensors](sensors/) to add your first sensor.
