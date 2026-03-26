#!/bin/bash
systemctl daemon-reload

# On full removal, clean up the virtualenv
if [ "$1" = "purge" ] || [ "$1" = "remove" ]; then
  rm -rf /usr/lib/temperature-sensor/venv
fi
exit 0
