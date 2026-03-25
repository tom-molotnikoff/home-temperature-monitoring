---
id: overview
title: Overview
sidebar_position: 1
slug: /
---

# Overview

Sensor Hub is a self-hosted home temperature monitoring system. It collects readings from physical temperature sensors deployed on your network, stores historical data, and presents it through a responsive web interface with real-time updates.

## Capabilities

- Collect temperature readings from DS18B20 sensors connected to Raspberry Pi devices over HTTP
- Store and visualize historical temperature data with hourly averages
- Monitor sensor health and connectivity status
- Configure alert rules that trigger when readings exceed defined thresholds
- Receive notifications in-app and via email when alerts fire
- Manage users, roles, and granular permissions
- Stream real-time data to connected clients over WebSocket
- Update system configuration at runtime through the UI or REST API

## Architecture

Sensor Hub is distributed as a single Go binary (`sensor-hub`) packaged as RPM and DEB. The binary contains:

- A Go backend that exposes a REST API and WebSocket server on port 8080. All API routes live under the `/api` prefix. It handles sensor polling, data storage, alerting, authentication, and session management.
- A React/TypeScript single-page application embedded into the binary via `//go:embed`. The binary serves both the API and the UI from `/`, so no separate web server or frontend build step is needed.
- An embedded SQLite database at `/var/lib/sensor-hub/sensor_hub.db` that stores all persistent data including sensor readings, user accounts, sessions, alert rules, and notifications. Database schema changes are managed by golang-migrate embedded migrations that run automatically on startup.

Nginx sits in front of the binary as a TLS reverse proxy, forwarding HTTPS traffic on port 443 to the Go process on `127.0.0.1:8080`. Nginx is installed and configured separately — see the [nginx setup guide](nginx-setup).

## How it works

Each temperature sensor runs a lightweight Flask API on a Raspberry Pi. The API reads the connected DS18B20 sensor over the 1-wire protocol and returns the current temperature as JSON.

Sensor Hub polls each registered sensor at a configurable interval (default: every 5 minutes). Readings are stored in SQLite and broadcast to connected UI clients via WebSocket. Hourly averages are computed and stored separately for efficient historical queries.

When a reading triggers an alert rule, the system generates a notification. Notifications are delivered in-app through the notification bell and, if configured, sent as emails via Gmail SMTP using OAuth 2.0.

## Get started

Review the [prerequisites](prerequisites) to prepare your environment, then follow the [installation guide](installation) to deploy Sensor Hub.
