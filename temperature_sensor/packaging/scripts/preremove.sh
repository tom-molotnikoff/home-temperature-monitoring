#!/bin/bash
# Stop and disable only on full removal (not upgrade)
# DEB: $1=remove on removal, $1=upgrade on upgrade
if [ "$1" = "remove" ]; then
  systemctl stop temperature-sensor || true
  systemctl disable temperature-sensor || true
fi
exit 0
