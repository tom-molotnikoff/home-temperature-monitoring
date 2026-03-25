#!/bin/bash
getent group sensor-hub >/dev/null || groupadd --system sensor-hub
getent passwd sensor-hub >/dev/null || useradd --system \
  --gid sensor-hub \
  --home-dir /var/lib/sensor-hub \
  --shell /usr/sbin/nologin \
  --comment "Sensor Hub service account" \
  sensor-hub
exit 0
