#!/bin/bash
getent group temperature-sensor >/dev/null || groupadd --system temperature-sensor
getent passwd temperature-sensor >/dev/null || useradd --system \
  --gid temperature-sensor \
  --home-dir /usr/lib/temperature-sensor \
  --shell /usr/sbin/nologin \
  --comment "Temperature Sensor service account" \
  temperature-sensor

# Add to gpio group for hardware access (1-wire, GPIO)
getent group gpio >/dev/null && usermod -aG gpio temperature-sensor || true
exit 0
