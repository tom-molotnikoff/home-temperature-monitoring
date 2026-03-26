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

### Check connectivity
```bash
sensor-hub health
```

### List all sensors
```bash
sensor-hub sensors list
```

### Get a specific sensor
```bash
sensor-hub sensors get "Living Room"
```

### Get temperature readings for a date range
```bash
sensor-hub readings between --from 2026-03-01 --to 2026-03-26

# Filter by sensor (name is case-sensitive, use quotes for names with spaces)
sensor-hub readings between --sensor "Living Room" --from 2026-03-01 --to 2026-03-26
```

### Get hourly averaged readings
```bash
sensor-hub readings hourly --sensor "Living Room" --from 2026-03-01 --to 2026-03-26
```

### Create an alert rule
```bash
sensor-hub alerts create --sensor-id 1 --type HIGH_TEMP --threshold 30
```

### Manage API keys
```bash
sensor-hub api-keys list
sensor-hub api-keys create --name "my-automation-key"
```
