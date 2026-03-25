#!/bin/bash
# Stop and disable only on full removal (not upgrade)
# RPM: $1=0 on removal, $1=1 on upgrade
# DEB: $1=remove on removal, $1=upgrade on upgrade
if [ "$1" = "0" ] || [ "$1" = "remove" ]; then
  systemctl stop sensor-hub || true
  systemctl disable sensor-hub || true
fi
exit 0
