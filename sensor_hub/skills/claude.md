---
name: sensor-hub
description: Interact with a Sensor Hub home temperature monitoring instance via CLI
---

# Sensor Hub CLI

`sensor-hub` is a CLI tool for interacting with a Sensor Hub instance — a home temperature monitoring system.

## Discovery

Run `sensor-hub --help` to see all available commands.
Run `sensor-hub <command> --help` for detailed usage of any command.
Run `sensor-hub <command> <subcommand> --help` for subcommand details.

## Prerequisites

The CLI must be configured with a server URL and API key:
- Config file: `~/.sensor-hub.yaml`
- Run `sensor-hub config init` for interactive setup
- Or pass `--server` and `--api-key` flags to any command

## Output

All commands output JSON to stdout. Errors go to stderr.
Pipe through `jq` for filtering: `sensor-hub sensors list | jq '.[].name'`

## Example Workflows

### Health
```bash
sensor-hub health
```

### Sensors
```bash
sensor-hub sensors list                              # List all sensors
sensor-hub sensors get "Living Room"                 # Get by name
sensor-hub sensors exists "Living Room"              # Check if exists
sensor-hub sensors list-by-type indoor               # List by type
sensor-hub sensors add --name X --type Y --url Z     # Create sensor
sensor-hub sensors update 1 --name X --type Y --url Z  # Update by ID
sensor-hub sensors delete "Living Room"              # Delete by name
sensor-hub sensors enable "Living Room"              # Enable sensor
sensor-hub sensors disable "Living Room"             # Disable sensor
sensor-hub sensors health "Living Room"              # Health history
sensor-hub sensors health "Living Room" --limit 10   # With limit
sensor-hub sensors stats                             # Total readings per sensor
sensor-hub sensors collect                           # Collect all
sensor-hub sensors collect "Living Room"             # Collect specific
```

### Readings
```bash
sensor-hub readings between --start 2026-03-01 --end 2026-03-26
sensor-hub readings between --sensor "Living Room" --start 2026-03-01 --end 2026-03-26
sensor-hub readings hourly --start 2026-03-01 --end 2026-03-26
sensor-hub readings hourly --sensor "Living Room" --start 2026-03-01 --end 2026-03-26
```

### Alerts
```bash
sensor-hub alerts list                               # List all rules
sensor-hub alerts get 1                              # Get by sensor ID
sensor-hub alerts create --sensor-id 1 --type HIGH_TEMP --threshold 30
sensor-hub alerts update 1 --alert-type HIGH_TEMP --high-threshold 30 --low-threshold 10 --enabled --rate-limit-hours 6
sensor-hub alerts delete 1                           # Delete by sensor ID
sensor-hub alerts history 1                          # Alert history
sensor-hub alerts history 1 --limit 20               # With limit
```

### Notifications
```bash
sensor-hub notifications list                        # List notifications
sensor-hub notifications list --limit 10 --offset 20 --include-dismissed
sensor-hub notifications unread-count                # Unread count
sensor-hub notifications read 5                      # Mark as read
sensor-hub notifications dismiss 5                   # Dismiss
sensor-hub notifications bulk-read                   # Mark all as read
sensor-hub notifications bulk-dismiss                # Dismiss all
sensor-hub notifications preferences                 # Get preferences
sensor-hub notifications set-preference --category threshold_alert --email-enabled --inapp-enabled
```

### Auth
```bash
sensor-hub auth login --username admin --password secret
sensor-hub auth logout
sensor-hub auth me                                   # Current user info
sensor-hub auth sessions                             # List sessions
sensor-hub auth revoke-session abc123                # Revoke session
```

### Users
```bash
sensor-hub users list
sensor-hub users create --username X --password Y --email Z
sensor-hub users delete 2
sensor-hub users change-password --user-id 1 --new-password newpass
sensor-hub users set-must-change 1 --must-change
sensor-hub users set-roles 1 --roles admin,viewer
```

### Roles
```bash
sensor-hub roles list                                # List roles
sensor-hub roles list-permissions                    # List all permissions
sensor-hub roles get-permissions 1                   # Permissions for role
sensor-hub roles assign-permission 1 --permission-id 5
sensor-hub roles remove-permission 1 5               # roleId permissionId
```

### API Keys
```bash
sensor-hub api-keys list
sensor-hub api-keys create --name "my-key"
sensor-hub api-keys update-expiry 3 --expires-at "2026-12-31T23:59:59Z"
sensor-hub api-keys revoke 3
sensor-hub api-keys delete 3
```

### OAuth
```bash
sensor-hub oauth status                              # Configuration status
sensor-hub oauth authorize                           # Start auth flow
sensor-hub oauth submit-code --code CODE --state STATE
sensor-hub oauth reload                              # Reload from disk
```

### Properties
```bash
sensor-hub properties get                            # Get all properties
sensor-hub properties set --key weather.latitude --value 53.3811
```

### Dashboards
```bash
sensor-hub dashboards list                           # List all dashboards
sensor-hub dashboards get 1                          # Get dashboard by ID
sensor-hub dashboards create --name "My Dashboard"   # Create dashboard
sensor-hub dashboards delete 1                       # Delete by ID
sensor-hub dashboards update 1 --file dashboard.json # Update from JSON file
```

The `update` command requires a JSON file with the full dashboard structure.

#### Dashboard JSON schema

```json
{
  "name": "My Dashboard",
  "config": {
    "widgets": [
      {
        "id": "unique-string-id",
        "type": "temperature-chart",
        "config": {},
        "layout": { "x": 0, "y": 0, "w": 6, "h": 4 }
      }
    ],
    "breakpoints": { "lg": 12, "md": 10, "sm": 6 }
  }
}
```

- `id`: Unique string per widget (e.g. UUID or descriptive slug)
- `layout`: Grid position — `x` (column), `y` (row), `w` (width in columns), `h` (height in row units)
- `breakpoints`: Columns per screen size (lg=large, md=medium, sm=small)

#### Available widget types and their config fields

| type                 | config fields                                                                               | description                                  |
|----------------------|---------------------------------------------------------------------------------------------|----------------------------------------------|
| `temperature-chart`  | `startDate` (date), `endDate` (date), `useHourlyAverages` (boolean)                         | Indoor temperature line chart                |
| `live-readings`      | —                                                                                           | Real-time temperature readings data grid     |
| `weather-forecast`   | —                                                                                           | External weather forecast card               |
| `sensor-health-pie`  | —                                                                                           | Sensor health status pie chart               |
| `sensor-type-pie`    | —                                                                                           | Sensor type distribution pie chart           |
| `health-timeline`    | `sensorId` (number), `limit` (number, default 1000)                                         | Sensor health status history chart           |
| `reading-stats`      | —                                                                                           | Total readings per sensor data grid          |
| `notifications-feed` | —                                                                                           | Recent notifications feed                    |
| `markdown-note`      | `content` (string)                                                                          | User-defined markdown text block             |
| `current-reading`    | `sensorId` (number)                                                                         | Big number display for a single sensor       |
| `min-max-avg`        | `sensorId` (number), `startDate` (date), `endDate` (date)                                   | Min/max/avg statistics for a sensor          |
| `gauge`              | `sensorId` (number), `min` (number, default 0), `max` (number, default 40)                  | Temperature gauge dial for a single sensor   |
| `comparison-chart`   | `sensorIds` (number[]), `startDate` (date), `endDate` (date), `useHourlyAverages` (boolean) | Multi-sensor overlay line chart              |
| `group-summary`      | —                                                                                           | Average temperature across all sensors       |
| `alert-summary`      | —                                                                                           | Compact list of configured alert rules       |
| `uptime`             | `sensorId` (number), `limit` (number, default 1000)                                         | Uptime percentage for a sensor               |
| `heatmap`            | `sensorId` (number)                                                                         | Color-coded 30-day temperature heatmap       |

**Config field notes:**
- `sensorId` is a numeric sensor ID (see `sensor-hub sensors list` to find IDs)
- `sensorIds` is an array of numeric sensor IDs
- `startDate` / `endDate` are ISO date strings (e.g. `"2026-04-01"`); if omitted, defaults to today→tomorrow
- `limit` controls how many history records to fetch; defaults to 1000 if omitted

#### Example: dashboard with two widgets

```json
{
  "name": "Living Room Monitor",
  "config": {
    "widgets": [
      {
        "id": "temp-chart-1",
        "type": "temperature-chart",
        "config": { "startDate": "2026-04-01", "endDate": "2026-04-03", "useHourlyAverages": true },
        "layout": { "x": 0, "y": 0, "w": 8, "h": 4 }
      },
      {
        "id": "gauge-living",
        "type": "gauge",
        "config": { "sensorId": 1, "min": 10, "max": 35 },
        "layout": { "x": 8, "y": 0, "w": 4, "h": 4 }
      }
    ],
    "breakpoints": { "lg": 12, "md": 10, "sm": 6 }
  }
}
```
