---
id: upgrading
title: Upgrading
sidebar_position: 4
---

# Upgrading

Sensor Hub uses Docker Compose for deployment and Flyway for database migrations. Upgrading involves pulling the latest code and rebuilding the containers.

## Upgrade process

1. Back up your MySQL data volume and the `configuration/` directory.

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

4. Flyway automatically detects and applies any new database migrations before the backend starts.

## Database migrations

Migrations are located in `sensor_hub/db/changesets/` and follow a sequential versioning scheme (V1, V2, and so on). Flyway tracks which migrations have been applied in the `flyway_schema_history` table and only runs new ones.

The `baselineOnMigrate=true` flag in the Flyway configuration ensures compatibility with existing databases that were set up before Flyway was introduced.

Migrations are forward-only. There is no automated rollback mechanism, which is why backing up the database before upgrading is recommended.

## Configuration changes

New releases may introduce additional configuration properties. These properties use sensible defaults, so existing configuration files continue to work without modification unless you want to override the new defaults.

Review release notes for any new properties and refer to [Configuration Settings](configuration) for the full property reference.

Configuration can also be updated at runtime through the web UI or API without restarting the services.
