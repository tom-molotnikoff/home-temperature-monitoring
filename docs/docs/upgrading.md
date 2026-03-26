---
id: upgrading
title: Upgrading
sidebar_position: 4
---

# Upgrading

## Back up the database

```bash
sudo cp /var/lib/sensor-hub/sensor_hub.db ~/sensor-hub-backup.db
```

## Download and install the new package

Download the latest package from the [GitHub Releases](https://github.com/tommolyit/home-temperature-monitoring/releases) page.

**Fedora / RHEL:**

```bash
sudo dnf upgrade ./sensor-hub-*.rpm
```

**Debian / Ubuntu:**

```bash
sudo apt install ./sensor-hub_*.deb
```

The postinstall scriptlet restarts the `sensor-hub.service` automatically.

## Database migrations

Embedded migrations run automatically on startup. The migrate library tracks which migrations have been applied in a `schema_migrations` table and only runs new ones. Migrations are embedded into the binary at build time using `//go:embed`, so no external migration tool is needed.

Migrations are forward-only. There is no automated rollback mechanism, which is why backing up the database before upgrading is recommended.

## Configuration files

Configuration files in `/etc/sensor-hub/` are marked as `noreplace` (RPM) or `conffiles` (DEB). Your edits are preserved during upgrades:

- **RPM:** If the package ships a new default, it is saved as `.rpmnew` alongside your existing file.
- **DEB:** If the package ships a new default, it is saved as `.dpkg-new` alongside your existing file.

Review release notes for any new properties and refer to [Configuration Settings](configuration) for the full property reference.

## Verify

```bash
sudo systemctl status sensor-hub
curl -k https://localhost/api/health
```
