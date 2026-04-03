---
id: cli-tool
title: CLI Tool
sidebar_position: 7
---

# CLI Tool

Sensor Hub ships as a single binary that can run as a **server** (`sensor-hub serve`) or as a **command-line client** for interacting with a remote Sensor Hub instance.

## Installation

### CLI-only package (recommended for remote machines)

A lightweight `sensor-hub-cli` package is available that contains just the binary and shell completions — no server, systemd service, or configuration files. Install it on any machine you want to manage Sensor Hub from:

**Fedora / RHEL:**

```bash
sudo dnf install ./sensor-hub-cli-*.rpm
```

**Debian / Ubuntu:**

```bash
sudo apt install ./sensor-hub-cli_*.deb
```

Download the latest package from the [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases) page. Packages are GPG-signed — see the [installation guide](installation) for verification steps.

:::note
The `sensor-hub-cli` and `sensor-hub` packages conflict with each other since they both provide the same binary. If you have the full server package installed, you already have the CLI — no need to install `sensor-hub-cli`.
:::

### Standalone binary

Alternatively, download a standalone binary from [GitHub Releases](https://github.com/tom-molotnikoff/home-temperature-monitoring/releases). Binaries are available for Linux, macOS, and Windows on both amd64 and arm64:

```bash
tar xzf sensor-hub_*_linux_amd64.tar.gz
sudo mv sensor-hub /usr/local/bin/sensor-hub
```

## Configuration

### Interactive Setup

Run the setup wizard to configure the CLI:

```bash
sensor-hub config init
```

This will prompt you for:

1. **Server URL** — the address of your Sensor Hub instance (e.g. `https://home.sensor-hub`)
2. **TLS verification** — if you entered an HTTPS URL, it asks whether to skip certificate verification (for self-signed certs)
3. **API key** — your API key for authentication

The wizard tests connectivity and API key validity before saving to `~/.sensor-hub.yaml`.

### Manual Configuration

Create `~/.sensor-hub.yaml`:

```yaml
server: https://home.sensor-hub
api_key: shk_your_api_key_here
insecure: true  # optional — skip TLS verification for self-signed certs
```

### Flag Overrides

All commands accept `--server`, `--api-key`, and `--insecure` flags, which override the config file:

```bash
sensor-hub sensors list --server https://home.sensor-hub --api-key shk_... --insecure
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

# Quick test without a config file
sensor-hub health --server https://home.sensor-hub --insecure
```

### Sensors

```bash
# List all sensors
sensor-hub sensors list

# Get a specific sensor by name
sensor-hub sensors get "Living Room"

# Check if a sensor exists
sensor-hub sensors exists "Living Room"

# List sensors by type
sensor-hub sensors list-by-type indoor

# Add a new sensor
sensor-hub sensors add --name "Bedroom" --type indoor --url http://192.168.1.50/api/temperature

# Update a sensor
sensor-hub sensors update 1 --name "Main Bedroom" --type indoor --url http://192.168.1.50/api/temperature

# Enable/disable a sensor
sensor-hub sensors enable "Bedroom"
sensor-hub sensors disable "Bedroom"

# Delete a sensor
sensor-hub sensors delete "Bedroom"

# View sensor health history
sensor-hub sensors health "Living Room"
sensor-hub sensors health "Living Room" --limit 10

# View total readings per sensor
sensor-hub sensors stats

# Trigger a data collection
sensor-hub sensors collect
sensor-hub sensors collect "Living Room"
```

### Readings

```bash
# Get readings between two dates
sensor-hub readings between --from 2025-01-01 --to 2025-01-31

# Filter by sensor
sensor-hub readings between --from 2025-01-01 --to 2025-01-31 --sensor "Living Room"

# Get hourly averages between two dates
sensor-hub readings hourly --from 2025-01-01 --to 2025-01-31

# Filter hourly averages by sensor
sensor-hub readings hourly --from 2025-01-01 --to 2025-01-31 --sensor "Living Room"
```

### Alerts

```bash
# List all alert rules
sensor-hub alerts list

# Get alert rules for a sensor
sensor-hub alerts get 1

# Create an alert rule
sensor-hub alerts create --sensor-id 1 --type HIGH_TEMP --threshold 30

# Update an alert rule
sensor-hub alerts update 1 --alert-type HIGH_TEMP --high-threshold 30 --low-threshold 10 --enabled --rate-limit-hours 6

# Delete alert rules for a sensor
sensor-hub alerts delete 1

# View alert history
sensor-hub alerts history 1
sensor-hub alerts history 1 --limit 20
```

### Notifications

```bash
# List notifications
sensor-hub notifications list
sensor-hub notifications list --limit 10 --offset 20 --include-dismissed

# Get unread count
sensor-hub notifications unread-count

# Mark a notification as read
sensor-hub notifications read 5

# Dismiss a notification
sensor-hub notifications dismiss 5

# Mark all notifications as read
sensor-hub notifications bulk-read

# Dismiss all notifications
sensor-hub notifications bulk-dismiss

# View notification preferences
sensor-hub notifications preferences

# Set a notification preference
sensor-hub notifications set-preference --category threshold_alert --email-enabled --inapp-enabled
```

### Auth

```bash
# Login with username and password
sensor-hub auth login --username admin --password secret123

# Get current user info
sensor-hub auth me

# List active sessions
sensor-hub auth sessions

# Revoke a session
sensor-hub auth revoke-session abc123

# Logout
sensor-hub auth logout
```

### Users

```bash
# List all users
sensor-hub users list

# Create a new user
sensor-hub users create --username admin2 --password secret123 --email admin2@example.com

# Delete a user
sensor-hub users delete 2

# Change a user's password
sensor-hub users change-password --user-id 1 --new-password newpass123

# Set whether a user must change password
sensor-hub users set-must-change 1 --must-change

# Set roles for a user
sensor-hub users set-roles 1 --roles admin,viewer
```

### Roles

```bash
# List all roles
sensor-hub roles list

# List all permissions
sensor-hub roles list-permissions

# Get permissions for a specific role
sensor-hub roles get-permissions 1

# Assign a permission to a role
sensor-hub roles assign-permission 1 --permission-id 5

# Remove a permission from a role
sensor-hub roles remove-permission 1 5
```

### API Keys

```bash
# List your API keys
sensor-hub api-keys list

# Create a new key
sensor-hub api-keys create --name "CI Pipeline"

# Update a key's expiry
sensor-hub api-keys update-expiry 3 --expires-at "2025-12-31T23:59:59Z"

# Revoke a key
sensor-hub api-keys revoke 3

# Delete a key
sensor-hub api-keys delete 3
```

### OAuth

```bash
# Check OAuth configuration status
sensor-hub oauth status

# Start authorization flow (returns auth URL + state token)
sensor-hub oauth authorize

# Submit authorization code from OAuth provider
sensor-hub oauth submit-code --code AUTH_CODE --state STATE_TOKEN

# Reload OAuth configuration from disk
sensor-hub oauth reload
```

### Properties

```bash
# Get all properties
sensor-hub properties get

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

### Dashboards

See the [Dashboards Guide](./dashboards) and [Dashboards API](./api/dashboards) for full details.

```bash
# List all dashboards
sensor-hub dashboards list

# Get a dashboard by ID (includes full widget and layout JSON)
sensor-hub dashboards get 1

# Create a new empty dashboard
sensor-hub dashboards create --name "My Dashboard"

# Update a dashboard from a JSON file
sensor-hub dashboards update 1 --file dashboard.json

# Delete a dashboard
sensor-hub dashboards delete 1
```

The `update` command expects a JSON file containing the full dashboard object (name, widgets, layouts). Use `get` to export an existing dashboard, edit the JSON, then `update` to apply changes.

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
