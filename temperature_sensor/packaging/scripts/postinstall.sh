#!/bin/bash
install -d -m 0750 -o temperature-sensor -g temperature-sensor /var/log/temperature-sensor

# Create virtualenv and install Python dependencies
INSTALL_DIR=/usr/lib/temperature-sensor
if [ ! -d "$INSTALL_DIR/venv" ]; then
  python3 -m venv "$INSTALL_DIR/venv"
fi
"$INSTALL_DIR/venv/bin/pip3" install --quiet --upgrade pip
"$INSTALL_DIR/venv/bin/pip3" install --quiet -r "$INSTALL_DIR/requirements.txt"

chown -R temperature-sensor:temperature-sensor "$INSTALL_DIR/venv"

systemctl daemon-reload

is_upgrade() {
  # DEB: $1=configure and $2 is the old version
  [ "$1" = "configure" ] && [ -n "$2" ] && return 0
  return 1
}

if is_upgrade "$@"; then
  # Re-install deps in case requirements changed
  systemctl restart temperature-sensor
else
  systemctl enable temperature-sensor
  echo ""
  echo "=========================================="
  echo " Temperature Sensor installed."
  echo " Configure: /etc/temperature-sensor/"
  echo " Then start: systemctl start temperature-sensor"
  echo "=========================================="
fi
exit 0
