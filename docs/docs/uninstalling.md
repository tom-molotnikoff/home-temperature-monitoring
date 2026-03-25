---
id: uninstalling
title: Uninstalling
sidebar_position: 5
---

# Uninstalling

## Remove the package

**Fedora / RHEL:**

```bash
sudo systemctl stop sensor-hub
sudo dnf remove sensor-hub
```

**Debian / Ubuntu:**

```bash
sudo systemctl stop sensor-hub
sudo apt remove sensor-hub
```

The package removal stops the service and removes the binary and systemd unit. Configuration files, data, and logs are preserved by default.

## Full cleanup (optional)

To remove all data, configuration, and logs:

```bash
sudo rm -rf /var/lib/sensor-hub /var/log/sensor-hub /etc/sensor-hub
```

To remove the system user and group:

```bash
sudo userdel sensor-hub
sudo groupdel sensor-hub
```
