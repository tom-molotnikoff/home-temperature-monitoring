# CLI Client & LLM Skills Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Restructure the `sensor-hub` binary into a dual-mode tool (server + CLI client) using Cobra, add API key authentication, and provide self-installing LLM skill files so AI assistants can interact with any Sensor Hub instance via the CLI.

**Architecture:** Refactor `main.go` from a flat `flag`-based server startup into a Cobra root command with subcommands. `sensor-hub serve` runs the existing server. All other subcommands are a lightweight HTTP/JSON client that reads `~/.sensor-hub.yaml` for connection details and authenticates via `X-API-Key` header. API keys are `shk_`-prefixed, shown once at creation, stored as SHA-256 hashes in a new `api_keys` table. A key inherits the creating user's role and permissions — no per-key permission scoping. Creating API keys requires the `manage_api_keys` permission assigned to a role. The binary also includes `sensor-hub skills install` to write LLM skill files into `.copilot/`, `.claude/`, etc.

**Tech Stack:** Go 1.25, Cobra CLI framework, Gin (existing), SQLite (existing), goreleaser (existing), Viper (config file reading)

---

## Task List (Outline)

### Phase 1: Foundation
1. **Add Cobra dependency** — `go get github.com/spf13/cobra` + `github.com/spf13/viper`
2. **Restructure main.go with Cobra** — root command, `serve` subcommand, migrate existing flags
3. **Update systemd service + packaging** — `ExecStart` → `sensor-hub serve`

### Phase 2: API Key Authentication (Backend)
4. **Database migration: api_keys table** — schema, migration file
5. **API key repository** — CRUD operations, hash storage
6. **API key service** — business logic, key generation, validation
7. **API key middleware** — `X-API-Key` header check, falls back to cookie auth
8. **API key REST endpoints** — create, list, update, revoke (UI + CLI accessible)

### Phase 3: API Key Management UI
9. **API key management UI page** — React page for creating, listing, editing, revoking keys

### Phase 4: CLI Client
10. **CLI config system** — `~/.sensor-hub.yaml`, `config init` wizard, `config show`
11. **CLI HTTP client** — shared client with auth, error handling, JSON output
12. **CLI sensor commands** — `sensors list`, `sensors get`, `sensors add`, `sensors delete`, `sensors enable/disable`
13. **CLI readings commands** — `readings between`, `readings hourly`
14. **CLI alerts commands** — `alerts list`, `alerts create`, `alerts delete`, `alerts history`
15. **CLI user/role commands** — `users list`, `users create`, `roles list`, `roles permissions`
16. **CLI properties commands** — `properties get`, `properties set`
17. **CLI notifications commands** — `notifications list`, `notifications read`, `notifications dismiss`
18. **CLI health command** — `health` (simple connectivity check)

### Phase 5: LLM Skills
19. **Embedded skill templates** — Go embed of skill markdown files for each LLM target
20. **Skills CLI commands** — `skills install --target copilot|claude|all`, `skills show`

### Phase 6: Build, Docs & Verification
21. **Update goreleaser** — multi-platform builds (linux, darwin, windows), update binary naming
22. **CLI documentation** — usage docs in `docs/` (Docusaurus)
23. **LLM skills documentation** — setup guide, example prompts, expected outputs
24. **Full verification** — go build, go test, tsc, npm run build

---

### Task 1: Add Cobra + Viper dependencies

**Files:**
- Modify: `sensor_hub/go.mod`

**Step 1: Add dependencies**

Run:
```bash
cd sensor_hub && go get github.com/spf13/cobra@latest github.com/spf13/viper@latest
```

**Step 2: Tidy**

Run: `cd sensor_hub && go mod tidy`

**Step 3: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 2: Restructure main.go with Cobra

**Files:**
- Modify: `sensor_hub/main.go` (gut and rebuild as Cobra root)
- Create: `sensor_hub/cmd/root.go` (root command with version flag)
- Create: `sensor_hub/cmd/serve.go` (move all existing server startup logic here)

**Step 1: Create cmd/root.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sensor-hub",
	Short: "Home temperature monitoring system",
	Long:  "Sensor Hub — a home temperature monitoring system.\nRun as a server with 'serve' or use CLI commands to interact with a remote instance.",
}

func Execute(version string) {
	Version = version
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 2: Create cmd/serve.go**

Move ALL existing `main()` logic (config init, DB init, service wiring, API listen) into a `runServe` function. The command definition:

```go
package cmd

import (
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Sensor Hub server",
	Long:  "Starts the HTTP API server, sensor discovery, periodic collection, and serves the embedded UI.",
	RunE:  runServe,
}

var configDir string
var logFile string

func init() {
	serveCmd.Flags().StringVar(&configDir, "config-dir", "configuration", "Path to configuration directory")
	serveCmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (default: stdout)")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// ... ALL existing main() logic from main.go, using configDir and logFile vars
	// Return error instead of log.Fatalf where possible
}
```

**Step 3: Rewrite main.go**

```go
package main

import "example/sensorHub/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
```

**Step 4: Verify**

Run: `cd sensor_hub && go build ./... && ./sensorHub serve --help`
Expected: Help text for `serve` command with `--config-dir` and `--log-file` flags.

Run: `cd sensor_hub && go test ./...`
Expected: All existing tests pass (server logic unchanged, just moved).

---

### Task 3: Update systemd service + packaging

**Files:**
- Modify: `packaging/sensor-hub.service` — change `ExecStart` to include `serve`
- Modify: `.air.toml` (if exists) — add `serve` to run command
- Modify: Docker compose files — add `serve` to entrypoint/command if binary is invoked directly

**Step 1: Update ExecStart**

In `packaging/sensor-hub.service`, change:
```
ExecStart=/usr/bin/sensor-hub --config-dir=/etc/sensor-hub --log-file=/var/log/sensor-hub/sensor-hub.log
```
To:
```
ExecStart=/usr/bin/sensor-hub serve --config-dir=/etc/sensor-hub --log-file=/var/log/sensor-hub/sensor-hub.log
```

**Step 2: Update AIR config (if it runs the binary directly)**

Check `.air.toml` in `sensor_hub/` — if it specifies a run command, add `serve` subcommand args.

**Step 3: Update docker-compose (docker_tests)**

Check `sensor_hub/docker_tests/docker-compose*.yml` — if the entrypoint runs the binary directly, add `serve`.

**Step 4: Verify**

Run: `cd sensor_hub && go build -o sensor-hub . && ./sensor-hub serve --help`
Expected: Shows serve help with flags.

Run: `./sensor-hub --help`
Expected: Shows root help listing `serve` and `version`.

---

### Task 4: Database migration — api_keys table

**Files:**
- Create: `sensor_hub/db/changesets/000002_api_keys.up.sql`
- Create: `sensor_hub/db/changesets/000002_api_keys.down.sql`

**Step 1: Write the up migration**

```sql
CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,           -- first 8 chars of the key (shk_xxxx) for identification
    key_hash TEXT NOT NULL UNIQUE,      -- SHA-256 hash of full key
    user_id INTEGER NOT NULL,           -- creating user (key inherits this user's role/permissions)
    expires_at DATETIME,                -- NULL = never expires
    revoked INTEGER NOT NULL DEFAULT 0, -- 0 = active, 1 = revoked
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
```

**Step 2: Write the down migration**

```sql
DROP TABLE IF EXISTS api_keys;
```

**Step 3: Verify migration runs**

Run: `cd sensor_hub && go test ./db/... -v`
Expected: Migration applies cleanly in test DB.

---

### Task 5: API key repository

**Files:**
- Create: `sensor_hub/db/api_key_repository.go`
- Create: `sensor_hub/db/api_key_repository_test.go`

**Step 1: Define the interface and struct**

```go
type ApiKey struct {
	Id         int        `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`   // "shk_xxxx" for display
	UserId     int        `json:"user_id"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Revoked    bool       `json:"revoked"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ApiKeyRepository interface {
	CreateApiKey(key ApiKey, keyHash string) (int64, error)
	GetApiKeyByHash(keyHash string) (*ApiKey, error)
	ListApiKeysForUser(userId int) ([]ApiKey, error)
	UpdateApiKeyExpiry(id int, expiresAt *time.Time) error
	RevokeApiKey(id int) error
	DeleteApiKey(id int) error
	UpdateLastUsed(id int) error
}
```

**Step 2: Implement SqlApiKeyRepository**

Follow the existing repository pattern (see `session_repository.go`). Key details:
- `GetApiKeyByHash` must check `revoked = 0` AND `(expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`
- `UpdateLastUsed` should be called on each successful auth (fire-and-forget, don't block request)

**Step 3: Write tests**

Test: create key, retrieve by hash, list for user, revoke, verify revoked key not returned, expired key not returned.

**Step 4: Verify**

Run: `cd sensor_hub && go test ./db/... -v -run ApiKey`

---

### Task 6: API key service

**Files:**
- Create: `sensor_hub/service/api_key_service.go`
- Create: `sensor_hub/service/api_key_service_test.go`

**Step 1: Define the service interface**

```go
type ApiKeyServiceInterface interface {
	CreateApiKey(name string, userId int) (fullKey string, err error)
	ListApiKeysForUser(userId int) ([]db.ApiKey, error)
	UpdateApiKeyExpiry(keyId int, userId int, expiresAt *time.Time) error
	RevokeApiKey(keyId int, userId int) error
	DeleteApiKey(keyId int, userId int) error
	ValidateApiKey(rawKey string) (*types.User, error)
}
```

**Step 2: Implement key generation**

- Generate 32 random bytes → base62 encode → prefix with `shk_`
- Full key: `shk_<44 chars>` (total ~48 chars)
- Hash: `SHA-256(fullKey)` stored in DB
- Return `fullKey` to caller ONCE — never stored or retrievable again

**Step 3: Implement ValidateApiKey**

1. Hash the raw key
2. Look up via `GetApiKeyByHash`
3. If found and valid: fetch the user via `userRepo`, populate their roles and permissions from the role system (same as session auth does)
4. Update `last_used_at` (async/fire-and-forget)
5. Return the fully-populated user object — downstream handlers see the same user as cookie auth

**Step 4: Write tests**

Test: key generation format, validation of valid/revoked/expired keys, returned user has correct role permissions.

**Step 5: Verify**

Run: `cd sensor_hub && go test ./service/... -v -run ApiKey`

---

### Task 7: API key middleware

**Files:**
- Modify: `sensor_hub/api/middleware/auth_middleware.go`

**Step 1: Add API key check to AuthRequired()**

Before the cookie check, add:

```go
apiKey := ctx.GetHeader("X-API-Key")
if apiKey != "" {
    user, err := apiKeyService.ValidateApiKey(apiKey)
    if err != nil || user == nil {
        ctx.AbortWithStatus(http.StatusUnauthorized)
        return
    }
    ctx.Set("currentUser", user)
    ctx.Set("authMethod", "api_key")
    ctx.Next()
    return
}
// ... existing cookie auth follows
```

**Step 2: Exempt API key auth from CSRF**

In `CSRFMiddleware()`, skip CSRF check if the `X-API-Key` header is present. API keys are not vulnerable to CSRF (they're not sent automatically by browsers).

**Step 3: Add init function**

```go
var apiKeyService service.ApiKeyServiceInterface

func InitApiKeyMiddleware(a service.ApiKeyServiceInterface) {
    apiKeyService = a
}
```

Wire this in `cmd/serve.go` alongside the existing `InitAuthMiddleware`.

**Step 4: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 8: API key REST endpoints

**Files:**
- Create: `sensor_hub/api/api_key_api.go`
- Modify: `sensor_hub/api/api.go` — register routes

**Step 1: Create api_key_api.go**

Endpoints (all under `/api/api-keys`, behind `AuthRequired()` + `RequirePermission("manage_api_keys")`):

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api-keys` | Create new key (returns full key ONCE in response) |
| GET | `/api-keys` | List current user's keys (prefix only, never full key) |
| PATCH | `/api-keys/:id/expiry` | Update key expiry |
| POST | `/api-keys/:id/revoke` | Revoke key (soft delete, POST not DELETE for clarity) |
| DELETE | `/api-keys/:id` | Hard delete key |

All handlers must verify the key belongs to the current user (ownership check).

**Step 2: Add `manage_api_keys` permission**

New migration `000003_api_key_permission.up.sql`:
```sql
INSERT INTO permissions (name, description) VALUES ('manage_api_keys', 'Create and manage API keys');
INSERT INTO role_permissions (role_id, permission_id) SELECT r.id, p.id FROM roles r, permissions p WHERE p.name = 'manage_api_keys';
```

**Step 3: Register routes in api.go**

Add `RegisterApiKeyRoutes(apiGroup)` alongside existing route registrations.

**Step 4: Wire service in cmd/serve.go**

Create `apiKeyRepo`, `apiKeyService`, call `api.InitApiKeyAPI(apiKeyService)` and `middleware.InitApiKeyMiddleware(apiKeyService)`.

**Step 5: Verify**

Run: `cd sensor_hub && go build ./... && go test ./...`

---

### Task 9: API key management UI page

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/api-keys/ApiKeysPage.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/ApiKeyDataGrid.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/CreateApiKeyDialog.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/hooks/useApiKeys.ts`
- Modify: Router config to add `/api-keys` route
- Modify: Navigation to add "API Keys" menu item

**Step 1: Create useApiKeys hook**

Fetches from `GET /api/api-keys`, returns `{ keys, loading, error, refetch }`. Standard REST hook pattern (see existing hooks for reference).

**Step 2: Create ApiKeyDataGrid**

MUI DataGrid showing: name, key prefix (`shk_xxxx...`), created date, last used date, expires at, status (active/revoked/expired). Actions column: edit expiry, revoke, delete.

**Step 3: Create CreateApiKeyDialog**

MUI Dialog with:
- Name field (text input)
- Optional expiry date picker
- On success: show the full key in a copyable alert with **strong warning** "This key will not be shown again"
- Copy-to-clipboard button
- Note: the key inherits the creating user's role/permissions automatically

**Step 4: Create ApiKeysPage**

Uses LayoutCard, TypographyH2, the DataGrid, and a "Create API Key" button that opens the dialog. Permission-gated to `manage_api_keys`.

**Step 5: Add route + navigation**

Add to router config and sidebar/nav menu.

**Step 6: Verify**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit && npx eslint . && npm run build`

---

### Task 10: CLI config system

**Files:**
- Create: `sensor_hub/cmd/config.go` — `config` command group
- Create: `sensor_hub/cmd/config_init.go` — interactive setup wizard
- Create: `sensor_hub/cmd/config_show.go` — show current config

**Step 1: Define config file structure**

`~/.sensor-hub.yaml`:
```yaml
server: https://my-sensor-hub.local:8080
api_key: shk_xxxxxxxxxxxxx
```

**Step 2: Create `config init` wizard**

Interactive prompts (use `bufio.Scanner` for stdin):
1. "Enter Sensor Hub server URL:" (default: http://localhost:8080)
2. "Enter API key:" (read, mask if possible)
3. Test connection: `GET /api/health` → show result
4. Test auth: `GET /api/auth/me` with API key → show username + permissions
5. Write `~/.sensor-hub.yaml` with `0600` permissions

**Step 3: Create `config show`**

Read and display current config (mask API key showing only prefix `shk_xxxx...`).

**Step 4: Create config loading helper**

Shared function `loadClientConfig()` used by all CLI commands:
1. Check `--server` and `--api-key` flags (override)
2. Fall back to `~/.sensor-hub.yaml` via Viper
3. Return `(serverURL, apiKey, error)`

Add persistent flags to root command:
```go
rootCmd.PersistentFlags().String("server", "", "Sensor Hub server URL (overrides config file)")
rootCmd.PersistentFlags().String("api-key", "", "API key (overrides config file)")
```

**Step 5: Verify**

Run: `cd sensor_hub && go build -o sensor-hub . && ./sensor-hub config init`

---

### Task 11: CLI HTTP client

**Files:**
- Create: `sensor_hub/cmd/client.go` — shared HTTP client with auth, error handling, JSON output

**Step 1: Implement client struct**

```go
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClientFromConfig(cmd *cobra.Command) (*Client, error) {
	// Load from flags → config file → error
}

func (c *Client) Get(path string, query url.Values) ([]byte, error)
func (c *Client) Post(path string, body interface{}) ([]byte, error)
func (c *Client) Patch(path string, body interface{}) ([]byte, error)
func (c *Client) Delete(path string) ([]byte, error)
```

**Step 2: Implement standard error handling**

- Non-2xx responses: print HTTP status + response body JSON to stderr, exit 1
- Connection errors: print "Error: could not connect to <url>" to stderr, exit 1
- All success output goes to stdout as JSON (pipeable)

**Step 3: Implement JSON output helper**

```go
func printJSON(data []byte) {
	// Pretty-print with indentation to stdout
	var buf bytes.Buffer
	json.Indent(&buf, data, "", "  ")
	fmt.Println(buf.String())
}
```

**Step 4: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 12: CLI sensor commands

**Files:**
- Create: `sensor_hub/cmd/sensors.go`

**Step 1: Implement sensor subcommands**

```
sensor-hub sensors list                        # GET /api/sensors/
sensor-hub sensors get <name>                  # GET /api/sensors/{name}
sensor-hub sensors add --name X --type Y --url Z  # POST /api/sensors/
sensor-hub sensors delete <name>               # DELETE /api/sensors/{name}
sensor-hub sensors enable <name>               # POST /api/sensors/enable/{name}
sensor-hub sensors disable <name>              # POST /api/sensors/disable/{name}
sensor-hub sensors health <name>               # GET /api/sensors/health/{name}
sensor-hub sensors stats                       # GET /api/sensors/stats/total-readings
sensor-hub sensors collect                     # POST /api/sensors/collect
sensor-hub sensors collect <name>              # POST /api/sensors/collect/{name}
```

Each command: create client → make request → print JSON → exit.

**Step 2: Verify**

Run: `cd sensor_hub && go build -o sensor-hub . && ./sensor-hub sensors --help`

---

### Task 13: CLI readings commands

**Files:**
- Create: `sensor_hub/cmd/readings.go`

**Step 1: Implement readings subcommands**

```
sensor-hub readings between --sensor X --from YYYY-MM-DD --to YYYY-MM-DD
sensor-hub readings hourly --sensor X --from YYYY-MM-DD --to YYYY-MM-DD
```

Maps to `GET /api/temperature/readings/between` and `/hourly/between` with query params.

**Step 2: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 14: CLI alerts commands

**Files:**
- Create: `sensor_hub/cmd/alerts.go`

**Step 1: Implement alerts subcommands**

```
sensor-hub alerts list                          # GET /api/alerts
sensor-hub alerts get <sensorId>                # GET /api/alerts/{sensorId}
sensor-hub alerts create --sensor-id X --type Y --threshold Z  # POST /api/alerts
sensor-hub alerts delete <sensorId>             # DELETE /api/alerts/{sensorId}
sensor-hub alerts history <sensorId>            # GET /api/alerts/{sensorId}/history
```

**Step 2: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 15: CLI user/role commands

**Files:**
- Create: `sensor_hub/cmd/users.go`
- Create: `sensor_hub/cmd/roles.go`

**Step 1: Implement user subcommands**

```
sensor-hub users list                           # GET /api/users
sensor-hub users get <id>                       # GET /api/users/{id}
sensor-hub users create --username X --password Y  # POST /api/users
sensor-hub users delete <id>                    # DELETE /api/users/{id}
sensor-hub users roles <id>                     # GET /api/users/{id}/roles
```

**Step 2: Implement role subcommands**

```
sensor-hub roles list                           # GET /api/roles
sensor-hub roles permissions                    # GET /api/roles/permissions
```

**Step 3: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 16: CLI properties commands

**Files:**
- Create: `sensor_hub/cmd/properties.go`

**Step 1: Implement properties subcommands**

```
sensor-hub properties get                       # GET /api/properties
sensor-hub properties set --key X --value Y     # PATCH /api/properties
```

**Step 2: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 17: CLI notifications commands

**Files:**
- Create: `sensor_hub/cmd/notifications.go`

**Step 1: Implement notifications subcommands**

```
sensor-hub notifications list                   # GET /api/notifications
sensor-hub notifications read <id>              # POST /api/notifications/{id}/read
sensor-hub notifications dismiss <id>           # POST /api/notifications/{id}/dismiss
sensor-hub notifications unread-count           # GET /api/notifications/unread-count
```

**Step 2: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 18: CLI health command

**Files:**
- Create: `sensor_hub/cmd/health.go`

**Step 1: Implement health check**

```
sensor-hub health                               # GET /api/health (no auth required)
```

Output: `{"status": "ok", "server": "https://...", "latency_ms": 42}`

No API key needed — this is also useful for connectivity testing during `config init`.

**Step 2: Verify**

Run: `cd sensor_hub && go build -o sensor-hub . && ./sensor-hub health --server http://localhost:8080`

---

### Task 19: Embedded skill templates

**Files:**
- Create: `sensor_hub/skills/copilot.md` — Copilot CLI skill template
- Create: `sensor_hub/skills/claude.md` — Claude Code skill template
- Create: `sensor_hub/skills/embed.go` — Go embed directive

**Step 1: Write the skill template**

The skill text should be minimal. It tells the LLM:
- What `sensor-hub` is (home temperature monitoring CLI)
- How to discover commands: `sensor-hub --help`, `sensor-hub <command> --help`
- How auth works: expects `~/.sensor-hub.yaml` to be configured
- Output is always JSON
- A few example workflows (check health, list sensors, get readings)

Example structure:
```markdown
---
name: sensor-hub
description: Interact with a Sensor Hub home temperature monitoring instance via CLI
---

# Sensor Hub CLI

`sensor-hub` is a CLI tool for interacting with a Sensor Hub instance.

## Discovery

Run `sensor-hub --help` to see all available commands.
Run `sensor-hub <command> --help` for detailed usage of any command.

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
sensor-hub health

### List all sensors
sensor-hub sensors list

### Get temperature readings for a sensor
sensor-hub readings between --sensor living-room --from 2026-03-01 --to 2026-03-26
```

**Step 2: Create embed.go**

```go
package skills

import "embed"

//go:embed *.md
var SkillFiles embed.FS
```

**Step 3: Verify**

Run: `cd sensor_hub && go build ./...`

---

### Task 20: Skills CLI commands

**Files:**
- Create: `sensor_hub/cmd/skills.go`

**Step 1: Implement skills subcommands**

```
sensor-hub skills show                          # Print skill text to stdout
sensor-hub skills install --target copilot      # Write to ~/.copilot/skills/sensor-hub/SKILL.md
sensor-hub skills install --target claude        # Write to ~/.claude/skills/sensor-hub/SKILL.md
sensor-hub skills install --all                  # Install to all detected LLM directories
```

**Step 2: Implement target detection**

For `--all`: check if `~/.copilot/`, `~/.claude/` exist. Install to each that exists. Report what was installed.

**Step 3: Implement installation**

1. Read embedded skill template from `skills.SkillFiles`
2. Create directory `~/.{target}/skills/sensor-hub/` if needed
3. Write `SKILL.md`
4. Print confirmation: "Installed sensor-hub skill to ~/.copilot/skills/sensor-hub/SKILL.md"

**Step 4: Verify**

Run: `cd sensor_hub && go build -o sensor-hub . && ./sensor-hub skills show`
Run: `./sensor-hub skills install --target copilot`

---

### Task 21: Update goreleaser for multi-platform

**Files:**
- Modify: `.goreleaser.yml`

**Step 1: Add platforms**

Change `goos` from `[linux]` to `[linux, darwin, windows]`.

This enables cross-compilation for macOS and Windows. The CLI client works on all platforms; `serve` is typically Linux-only but there's no reason to restrict the binary.

**Step 2: Verify**

Run: `cd sensor_hub && GOOS=darwin GOARCH=arm64 go build ./...` (cross-compile check)
Run: `cd sensor_hub && GOOS=windows GOARCH=amd64 go build ./...`

---

### Task 22: CLI documentation

**Files:**
- Create: `docs/docs/cli/installation.md`
- Create: `docs/docs/cli/configuration.md`
- Create: `docs/docs/cli/commands.md`

**Step 1: Installation docs**

Cover: download binary from releases, RPM/DEB install, verify with `sensor-hub --version`.

**Step 2: Configuration docs**

Cover: `sensor-hub config init` walkthrough, manual `~/.sensor-hub.yaml` creation, flag overrides, multiple instances.

**Step 3: Commands reference**

Auto-generate from `sensor-hub --help` output, or manually document each command group with examples and expected output.

**Step 4: Verify**

Run: `cd docs && npm run build` (Docusaurus build)

---

### Task 23: LLM skills documentation

**Files:**
- Create: `docs/docs/cli/llm-skills.md`

**Step 1: Write the guide**

Cover:
- What LLM skills are and why they're useful
- `sensor-hub skills install --target copilot` walkthrough
- `sensor-hub skills install --target claude` walkthrough
- `sensor-hub skills install --all` for convenience
- `sensor-hub skills show` for manual installation
- Example prompts and expected outputs:
  - "Check if my sensor hub is healthy"
  - "Show me all registered sensors"
  - "Get temperature readings from the living room sensor for the past week"
  - "Create an alert for when the bedroom temperature drops below 15°C"

**Step 2: Verify**

Run: `cd docs && npm run build`

---

### Task 24: Full verification

**Step 1: Go build + test**

```bash
cd sensor_hub && go build ./... && go test ./...
```

**Step 2: UI build**

```bash
cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit && npx eslint . && npm run build
```

**Step 3: Docs build**

```bash
cd docs && npm run build
```

**Step 4: CLI smoke test**

```bash
./sensor-hub --help
./sensor-hub serve --help
./sensor-hub config init --help
./sensor-hub sensors --help
./sensor-hub skills show
./sensor-hub skills install --all
```

**Step 5: Cross-compile**

```bash
GOOS=darwin GOARCH=arm64 go build -o /dev/null ./...
GOOS=windows GOARCH=amd64 go build -o /dev/null ./...
```
