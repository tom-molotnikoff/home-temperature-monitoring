#!/bin/bash
install -d -m 0750 -o sensor-hub -g sensor-hub /var/lib/sensor-hub
install -d -m 0750 -o sensor-hub -g sensor-hub /var/log/sensor-hub

systemctl daemon-reload

is_upgrade() {
  # RPM: $1=2 on upgrade
  [ "$1" = "2" ] && return 0
  # DEB: $1=configure and $2 is the old version
  [ "$1" = "configure" ] && [ -n "$2" ] && return 0
  return 1
}

if is_upgrade "$@"; then
  systemctl restart sensor-hub
else
  systemctl enable sensor-hub
  echo ""
  echo "=========================================="
  echo " Sensor Hub installed successfully."
  echo " Configure: /etc/sensor-hub/"
  echo " Then start: systemctl start sensor-hub"
  echo "=========================================="
fi
exit 0
