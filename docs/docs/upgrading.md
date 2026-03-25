---
id: upgrading
title: Upgrading
sidebar_position: 4
---

# Upgrading

Sensor Hub uses Docker Compose for deployment with embedded database migrations. Upgrading involves pulling the latest code and rebuilding the containers.

## Upgrade process

1. Back up your SQLite database file and the `configuration/` directory.

2. Pull the latest changes:

```bash
cd home-temperature-monitoring
git pull
```

3. Rebuild and restart the containers:

```bash
cd sensor_hub/docker
docker compose up -d --build
```

4. Embedded migrations are applied automatically on startup.

## Database migrations

Migrations are located in `sensor_hub/db/migrations/` and follow the golang-migrate naming convention (e.g., `000001_init_schema.up.sql`). The migrate library tracks which migrations have been applied in a `schema_migrations` table and only runs new ones. Migrations are embedded into the binary at build time using `//go:embed`, so no external migration tool is needed.

Migrations are forward-only. There is no automated rollback mechanism, which is why backing up the database before upgrading is recommended.

## Configuration changes

New releases may introduce additional configuration properties. These properties use sensible defaults, so existing configuration files continue to work without modification unless you want to override the new defaults.

Review release notes for any new properties and refer to [Configuration Settings](configuration) for the full property reference.

Configuration can also be updated at runtime through the web UI or API without restarting the services.
