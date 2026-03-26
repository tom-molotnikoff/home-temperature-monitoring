---
id: cli
title: CLI Tool
sidebar_position: 7
---

# CLI Tool

Sensor Hub ships as a single binary that can run as a **server** (`sensor-hub serve`) or as a **command-line client** for interacting with a remote Sensor Hub instance.

## Installation

Download the latest release for your platform from the [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases) page. Binaries are available for Linux, macOS, and Windows on both amd64 and arm64 architectures.

```bash
# Linux (amd64)
chmod +x sensor-hub_linux_amd64
sudo mv sensor-hub_linux_amd64 /usr/local/bin/sensor-hub

# macOS (Apple Silicon)
chmod +x sensor-hub_darwin_arm64
sudo mv sensor-hub_darwin_arm64 /usr/local/bin/sensor-hub
```

If you installed Sensor Hub via a package manager (RPM/DEB), the binary is already at `/usr/bin/sensor-hub`.

## Configuration

### Interactive Setup

Run the setup wizard to configure the CLI:

```bash
sensor-hub config init
```

This will prompt you for your server URL and API key, test connectivity, and save the configuration to `~/.sensor-hub.yaml`.

### Manual Configuration

Create `~/.sensor-hub.yaml`:

```yaml
server: https://sensors.example.com
api_key: shk_your_api_key_here
```

### Flag Overrides

All commands accept `--server` and `--api-key` flags, which override the config file:

```bash
sensor-hub sensors list --server https://sensors.example.com --api-key shk_...
```

### View Current Configuration

```bash
sensor-hub config show
```

## Creating an API Key

You need an API key to authenticate CLI requests. Create one from the web UI (**API Keys** in the sidebar) or using the CLI itself if you already have one:

```bash
sensor-hub api-keys create --name "My CLI Key"
```

:::warning
The full API key is shown only once at creation time. Store it securely.
:::

## Commands

### Health Check

Test connectivity to the server (no authentication required):

```bash
sensor-hub health
```

### Sensors

```bash
# List all sensors
sensor-hub sensors list

# Get a specific sensor
sensor-hub sensors get --name "Living Room"

# Add a new sensor
sensor-hub sensors add --name "Bedroom" --type Temperature --url http://192.168.1.50/api/temperature

# Enable/disable a sensor
sensor-hub sensors enable --name "Bedroom"
sensor-hub sensors disable --name "Bedroom"

# Delete a sensor
sensor-hub sensors delete --name "Bedroom"

# View sensor health status
sensor-hub sensors health

# View total readings per sensor
sensor-hub sensors stats

# Trigger a data collection
sensor-hub sensors collect
sensor-hub sensors collect --name "Living Room"
```

### Readings

```bash
# Get readings between two dates
sensor-hub readings between --from 2025-01-01 --to 2025-01-31

# Filter by sensor
sensor-hub readings between --from 2025-01-01 --to 2025-01-31 --sensor "Living Room"

# Get hourly averages for a specific date
sensor-hub readings hourly --date 2025-01-15

# Filter hourly averages by sensor
sensor-hub readings hourly --date 2025-01-15 --sensor "Living Room"
```

### Alerts

```bash
# List all alert rules
sensor-hub alerts list

# Get a specific alert rule
sensor-hub alerts get --id 1

# Create a numeric range alert
sensor-hub alerts create --sensor-id 1 --type numeric_range --high 30 --low 10

# Create a status-based alert
sensor-hub alerts create --sensor-id 1 --type status_based --trigger-status bad

# Delete an alert rule
sensor-hub alerts delete --id 1

# View alert history
sensor-hub alerts history
```

### Notifications

```bash
# List notifications
sensor-hub notifications list

# Get unread count
sensor-hub notifications unread-count

# Mark a notification as read
sensor-hub notifications read --id 5

# Dismiss a notification
sensor-hub notifications dismiss --id 5
```

### Users & Roles

```bash
# List all users
sensor-hub users list

# Get a specific user
sensor-hub users get --id 1

# Create a new user
sensor-hub users create --username admin2 --password secret123 --role-id 1

# Delete a user
sensor-hub users delete --id 2

# List roles
sensor-hub roles list

# List permissions for a role
sensor-hub roles permissions --id 1
```

### API Keys

```bash
# List your API keys
sensor-hub api-keys list

# Create a new key
sensor-hub api-keys create --name "CI Pipeline"

# Create a key with expiry
sensor-hub api-keys create --name "Temp Key" --expires "2025-12-31T23:59:59Z"

# Revoke a key
sensor-hub api-keys revoke --id 3

# Delete a key
sensor-hub api-keys delete --id 3
```

### Properties

```bash
# Get a property value
sensor-hub properties get --key weather.latitude

# Set a property value
sensor-hub properties set --key weather.latitude --value 53.3811
```

### Skills (LLM Integration)

See the [LLM Skills Guide](./llm-skills) for full details.

```bash
# Show a skill file
sensor-hub skills show --target copilot

# Install skill files for all supported LLM tools
sensor-hub skills install --all

# Install for a specific target
sensor-hub skills install --target claude
```

## Output Format

All commands output JSON to stdout, making it easy to pipe into `jq` or other tools:

```bash
# Pretty-print sensor list
sensor-hub sensors list | jq '.[] | {name, healthStatus}'

# Count active sensors
sensor-hub sensors list | jq '[.[] | select(.enabled == true)] | length'
```

Errors are written to stderr, so stdout always contains clean JSON output.

## Shell Completion

Sensor Hub supports shell completion for Bash, Zsh, Fish, and PowerShell:

```bash
# Bash
sensor-hub completion bash > /etc/bash_completion.d/sensor-hub

# Zsh
sensor-hub completion zsh > "${fpath[1]}/_sensor-hub"

# Fish
sensor-hub completion fish > ~/.config/fish/completions/sensor-hub.fish
```
