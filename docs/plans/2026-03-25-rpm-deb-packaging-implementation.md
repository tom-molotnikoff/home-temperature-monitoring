# RPM/DEB Packaging Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Package sensor-hub as signed RPM and DEB packages with FHS-compliant layout, systemd integration, and automated GitHub Actions release pipeline.

**Architecture:** Add `--config-dir`, `--log-file`, and `--version` CLI flags to the Go binary. Create packaging scaffolding (systemd unit, logrotate, scriptlets, goreleaser config). Set up a GitHub Action that builds the React UI, cross-compiles for amd64+arm64, packages as RPM+DEB, signs with GPG, and publishes to GitHub Releases. Overhaul documentation to reflect package-based deployment.

**Tech Stack:** goreleaser (build + package + release), nfpm (RPM/DEB generation, used by goreleaser), GPG (package signing), GitHub Actions (CI/CD)

**Design Doc:** `docs/plans/2026-03-25-rpm-deb-packaging-design.md`

---

## Task List

1. **Add CLI flags to main.go** — `--config-dir`, `--log-file`, `--version`
2. **Make config loading accept configurable directory** — Update `application_properties` package
3. **Update OAuth path defaults** — Use config dir for credentials.json/token.json
4. **Update tests** — Ensure existing tests pass with refactored config loading
5. **Create packaging scaffolding** — systemd unit, logrotate, scriptlets, environment file, nginx example, production defaults
6. **Create goreleaser configuration** — `.goreleaser.yml` with nfpm, cross-compilation, signing
7. **Create GitHub Actions release workflow** — `.github/workflows/release.yml`
8. **Create local build script** — `scripts/build-packages.sh`
9. **PAUSE: GPG key setup** — Instructions for the user to generate keys and add to GitHub secrets
10. **Rewrite user-facing documentation** — prerequisites, installation, configuration, upgrading, uninstalling, nginx setup
11. **Create developer documentation** — building from source, Docker dev environment, releasing, testing
12. **Update READMEs** — Root README and sensor_hub README
13. **Final verification** — Build, test, local package build

---

### Task 1: Add CLI flags to main.go

**Files:**
- Modify: `sensor_hub/main.go`

**Step 1: Add version variable and flag parsing**

Add a `var version` string (set via ldflags) and use Go's `flag` package to parse `--config-dir`, `--log-file`, and `--version`.

```go
package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    // ... existing imports
)

var version = "dev"

func main() {
    configDir := flag.String("config-dir", "configuration", "Path to configuration directory")
    logFile := flag.String("log-file", "", "Path to log file (default: stdout)")
    showVersion := flag.Bool("version", false, "Print version and exit")
    flag.Parse()

    if *showVersion {
        fmt.Printf("sensor-hub %s\n", version)
        os.Exit(0)
    }

    // Set up logging
    log.SetPrefix("sensor-hub: ")
    if *logFile != "" {
        f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            log.Fatalf("failed to open log file: %v", err)
        }
        defer f.Close()
        mw := io.MultiWriter(os.Stdout, f)
        log.SetOutput(mw)
    }

    // Pass configDir to InitialiseConfig
    err := appProps.InitialiseConfig(*configDir)
    // ... rest unchanged
}
```

**Step 2: Verify it compiles**

Run: `cd sensor_hub && go build ./...`
Expected: Build will fail because `InitialiseConfig` doesn't accept a parameter yet — that's Task 2.

---

### Task 2: Make config loading accept configurable directory

**Files:**
- Modify: `sensor_hub/application_properties/application_properties_files.go`
- Modify: `sensor_hub/application_properties/application_configuration.go`
- Modify: `sensor_hub/application_properties/application_properties_defaults.go`

**Step 1: Add SetConfigDir function to application_properties_files.go**

Change the hardcoded file paths to be set via a `setConfigPaths` function. The variable defaults remain `"configuration/..."` for backward compatibility (Docker, tests).

```go
// Replace lines 15-17:
var applicationPropertiesFilePath = "configuration/application.properties"
var smtpPropertiesFilePath = "configuration/smtp.properties"
var databasePropertiesFilePath = "configuration/database.properties"

// With:
var configDir = "configuration"
var applicationPropertiesFilePath string
var smtpPropertiesFilePath string
var databasePropertiesFilePath string

func init() {
    setConfigPaths(configDir)
}

func setConfigPaths(dir string) {
    configDir = dir
    applicationPropertiesFilePath = filepath.Join(dir, "application.properties")
    smtpPropertiesFilePath = filepath.Join(dir, "smtp.properties")
    databasePropertiesFilePath = filepath.Join(dir, "database.properties")
}
```

Add `"path/filepath"` to the imports.

**Step 2: Update InitialiseConfig to accept configDir parameter**

In `application_configuration.go`, change:

```go
// From:
func InitialiseConfig() error {

// To:
func InitialiseConfig(dir string) error {
    setConfigPaths(dir)
```

**Step 3: Update OAuth path defaults**

In `application_properties_defaults.go`, change the OAuth defaults to just filenames (not paths):
```go
"oauth.credentials.file.path": "credentials.json",
"oauth.token.file.path":       "token.json",
```

Then in `LoadConfigurationFromMaps` in `application_configuration.go`, resolve relative paths against `configDir`:
```go
if v, ok := appProps["oauth.credentials.file.path"]; ok {
    if !filepath.IsAbs(v) {
        v = filepath.Join(configDir, v)
    }
    cfg.OAuthCredentialsFilePath = v
}
if v, ok := appProps["oauth.token.file.path"]; ok {
    if !filepath.IsAbs(v) {
        v = filepath.Join(configDir, v)
    }
    cfg.OAuthTokenFilePath = v
}
```

**Step 4: Add configDir getter for other packages**

```go
func GetConfigDir() string {
    return configDir
}
```

**Step 5: Build and verify**

Run: `cd sensor_hub && go build ./...`
Expected: PASS

---

### Task 3: Update OAuth hardcoded fallback paths

**Files:**
- Modify: `sensor_hub/oauth/oauth.go`

**Step 1: Remove hardcoded fallback paths**

Find the hardcoded `"configuration/credentials.json"` and `"configuration/token.json"` fallbacks and replace them with dynamic paths:

```go
credPath := cfg.OAuthCredentialsFilePath
if credPath == "" {
    credPath = filepath.Join(appProps.GetConfigDir(), "credentials.json")
}
tokenPath := cfg.OAuthTokenFilePath
if tokenPath == "" {
    tokenPath = filepath.Join(appProps.GetConfigDir(), "token.json")
}
```

Add imports for `"path/filepath"` and the `appProps` alias.

**Step 2: Build and verify**

Run: `cd sensor_hub && go build ./...`
Expected: PASS

---

### Task 4: Verify all tests pass

**Files:** None (verification only)

**Step 1: Run full test suite**

Run: `cd sensor_hub && go test ./...`
Expected: All 11 packages PASS. The config changes are backward-compatible (default configDir is still `"configuration"`).

**Step 2: Verify --version flag works**

Run: `cd sensor_hub && go run . --version`
Expected: `sensor-hub dev`

---

### Task 5: Create packaging scaffolding

**Files:**
- Create: `packaging/sensor-hub.service`
- Create: `packaging/logrotate.conf`
- Create: `packaging/nginx.conf.example`
- Create: `packaging/environment`
- Create: `packaging/defaults/application.properties`
- Create: `packaging/defaults/database.properties`
- Create: `packaging/defaults/smtp.properties`
- Create: `packaging/scripts/preinstall.sh`
- Create: `packaging/scripts/postinstall.sh`
- Create: `packaging/scripts/preremove.sh`
- Create: `packaging/scripts/postremove.sh`

All file contents are specified in the design doc: `docs/plans/2026-03-25-rpm-deb-packaging-design.md`

**Step 1: Create directory structure**

```bash
mkdir -p packaging/{defaults,scripts}
```

**Step 2: Create systemd unit file** (`packaging/sensor-hub.service`)

```ini
[Unit]
Description=Sensor Hub - Home Temperature Monitoring
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=sensor-hub
Group=sensor-hub
ExecStart=/usr/bin/sensor-hub --config-dir=/etc/sensor-hub --log-file=/var/log/sensor-hub/sensor-hub.log
WorkingDirectory=/var/lib/sensor-hub
Restart=on-failure
RestartSec=5

NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/sensor-hub /var/log/sensor-hub
PrivateTmp=yes

Environment=SENSOR_HUB_PRODUCTION=true
EnvironmentFile=-/etc/sensor-hub/environment

[Install]
WantedBy=multi-user.target
```

**Step 3: Create logrotate config** (`packaging/logrotate.conf`)

```
/var/log/sensor-hub/sensor-hub.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    maxsize 50M
    copytruncate
}
```

**Step 4: Create nginx example** (`packaging/nginx.conf.example`)

Copy from `sensor_hub/ui/sensor_hub_ui/nginx/default.conf` — it's already the simplified TLS proxy config.

**Step 5: Create environment file template** (`packaging/environment`)

```bash
# Environment variables for sensor-hub systemd service.
# Uncomment and set values as needed.

# Create an initial admin user on first run (format: username:password)
# SENSOR_HUB_INITIAL_ADMIN=admin:changeme

# Enable TLS on the Go binary (usually not needed — nginx handles TLS)
# TLS_CERT_FILE=/etc/ssl/certs/sensor-hub.pem
# TLS_KEY_FILE=/etc/ssl/private/sensor-hub-key.pem

# Allow CORS from a specific origin (only needed for separate UI dev server)
# SENSOR_HUB_ALLOWED_ORIGIN=http://localhost:3000
```

**Step 6: Create production default configs** (`packaging/defaults/`)

- `application.properties` — Copy from `sensor_hub/configuration/application.properties` but set `sensor.discovery.skip=true` and `openapi.yaml.location=` (empty, since sensor discovery uses it but is skipped by default).
- `database.properties` — Set `database.path=/var/lib/sensor-hub/sensor_hub.db`
- `smtp.properties` — Copy from `sensor_hub/configuration/smtp.properties` (empty smtp.user)

**Step 7: Create scriptlets**

`packaging/scripts/preinstall.sh`:
```bash
#!/bin/bash
getent group sensor-hub >/dev/null || groupadd --system sensor-hub
getent passwd sensor-hub >/dev/null || useradd --system \
  --gid sensor-hub \
  --home-dir /var/lib/sensor-hub \
  --shell /usr/sbin/nologin \
  --comment "Sensor Hub service account" \
  sensor-hub
exit 0
```

`packaging/scripts/postinstall.sh`:
```bash
#!/bin/bash
install -d -m 0750 -o sensor-hub -g sensor-hub /var/lib/sensor-hub
install -d -m 0750 -o sensor-hub -g sensor-hub /var/log/sensor-hub
systemctl daemon-reload

# Fresh install ($1=1 on RPM, $1=configure on DEB)
if [ "$1" = "1" ] || [ "$1" = "configure" ]; then
  systemctl enable sensor-hub
  echo ""
  echo "=========================================="
  echo " Sensor Hub installed successfully."
  echo " Configure: /etc/sensor-hub/"
  echo " Then start: systemctl start sensor-hub"
  echo "=========================================="
fi

# Upgrade ($1=2 on RPM)
if [ "$1" = "2" ]; then
  systemctl restart sensor-hub
fi
exit 0
```

`packaging/scripts/preremove.sh`:
```bash
#!/bin/bash
if [ "$1" = "0" ] || [ "$1" = "remove" ]; then
  systemctl stop sensor-hub || true
  systemctl disable sensor-hub || true
fi
exit 0
```

`packaging/scripts/postremove.sh`:
```bash
#!/bin/bash
systemctl daemon-reload
exit 0
```

**Step 8: Verify all files are in place**

```bash
find packaging/ -type f | sort
```

Expected:
```
packaging/defaults/application.properties
packaging/defaults/database.properties
packaging/defaults/smtp.properties
packaging/environment
packaging/logrotate.conf
packaging/nginx.conf.example
packaging/scripts/postinstall.sh
packaging/scripts/postremove.sh
packaging/scripts/preinstall.sh
packaging/scripts/preremove.sh
packaging/sensor-hub.service
```

---

### Task 6: Create goreleaser configuration

**Files:**
- Create: `.goreleaser.yml` (repo root)

**Step 1: Create `.goreleaser.yml`**

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: sensor-hub
    dir: sensor_hub
    binary: sensor-hub
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - formats:
      - tar.gz
    name_template: >-
      sensor-hub_{{ .Version }}_{{ .Os }}_{{ .Arch }}

nfpms:
  - id: sensor-hub
    package_name: sensor-hub
    vendor: Home Temperature Monitoring
    homepage: https://github.com/YOUR_USER/home-temperature-monitoring
    maintainer: Tom Molotnikoff
    description: Home temperature monitoring system with embedded web UI
    license: MIT
    formats:
      - rpm
      - deb
    bindir: /usr/bin

    # Config files (noreplace — user edits survive upgrades)
    contents:
      - src: packaging/sensor-hub.service
        dst: /usr/lib/systemd/system/sensor-hub.service
      - src: packaging/logrotate.conf
        dst: /etc/logrotate.d/sensor-hub
      - src: packaging/nginx.conf.example
        dst: /etc/sensor-hub/nginx.conf.example
      - src: packaging/environment
        dst: /etc/sensor-hub/environment
        type: config|noreplace
        file_info:
          mode: 0640
          owner: root
          group: sensor-hub
      - src: packaging/defaults/application.properties
        dst: /etc/sensor-hub/application.properties
        type: config|noreplace
        file_info:
          mode: 0640
          owner: root
          group: sensor-hub
      - src: packaging/defaults/database.properties
        dst: /etc/sensor-hub/database.properties
        type: config|noreplace
        file_info:
          mode: 0640
          owner: root
          group: sensor-hub
      - src: packaging/defaults/smtp.properties
        dst: /etc/sensor-hub/smtp.properties
        type: config|noreplace
        file_info:
          mode: 0640
          owner: root
          group: sensor-hub

    scripts:
      preinstall: packaging/scripts/preinstall.sh
      postinstall: packaging/scripts/postinstall.sh
      preremove: packaging/scripts/preremove.sh
      postremove: packaging/scripts/postremove.sh

    rpm:
      group: Applications/System

    overrides:
      deb:
        dependencies:
          - adduser

signs:
  - artifacts: all
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
```

**Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.goreleaser.yml'))"`
Expected: No error

---

### Task 7: Create GitHub Actions release workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Step 1: Create workflow directory**

```bash
mkdir -p .github/workflows
```

**Step 2: Create release.yml**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: "25"
          cache: "npm"
          cache-dependency-path: sensor_hub/ui/sensor_hub_ui/package-lock.json

      - name: Build React UI
        working-directory: sensor_hub/ui/sensor_hub_ui
        run: |
          npm ci
          npm run build

      - name: Copy UI dist to web/dist
        run: cp -r sensor_hub/ui/sensor_hub_ui/dist/* sensor_hub/web/dist/

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: sensor_hub/go.mod
          cache-dependency-path: sensor_hub/go.sum

      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.GPG_PASSPHRASE }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

Note: The GPG fingerprint step needs the `id: import_gpg` attribute on the Import GPG key step. Update accordingly.

---

### Task 8: Create local build script

**Files:**
- Create: `scripts/build-packages.sh`

**Step 1: Create the script**

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==> Building React UI..."
cd "$REPO_ROOT/sensor_hub/ui/sensor_hub_ui"
npm ci --silent
npm run build

echo "==> Copying UI dist..."
cp -r dist/* "$REPO_ROOT/sensor_hub/web/dist/"

echo "==> Building packages with goreleaser (snapshot, unsigned)..."
cd "$REPO_ROOT"
goreleaser release --snapshot --skip=publish --skip=sign --clean

echo ""
echo "==> Packages built successfully:"
ls -lh dist/*.rpm dist/*.deb 2>/dev/null || echo "(no packages found — check dist/)"
```

**Step 2: Make executable**

```bash
chmod +x scripts/build-packages.sh
```

---

### Task 9: PAUSE — GPG key setup

**This is a manual step for the user.** Stop execution and provide instructions.

**Instructions for the user:**

1. **Generate a GPG keypair** (on your local machine):
   ```bash
   gpg --full-generate-key
   # Choose: RSA and RSA, 4096 bits, no expiration (or your preference)
   # Name: Sensor Hub Release Signing
   # Email: your-email@example.com
   ```

2. **Get the fingerprint:**
   ```bash
   gpg --list-secret-keys --keyid-format=long
   # Note the fingerprint (40-char hex string)
   ```

3. **Export the private key:**
   ```bash
   gpg --armor --export-secret-keys YOUR_FINGERPRINT > sensor-hub-gpg-private.key
   ```

4. **Add to GitHub Secrets** (Settings → Secrets and variables → Actions):
   - `GPG_PRIVATE_KEY` — paste the contents of `sensor-hub-gpg-private.key`
   - `GPG_PASSPHRASE` — the passphrase you used when generating the key

5. **Delete the exported key file:**
   ```bash
   rm sensor-hub-gpg-private.key
   ```

6. **(Optional) Publish your public key** so users can verify packages:
   ```bash
   gpg --armor --export YOUR_FINGERPRINT > sensor-hub-gpg-public.key
   # Add this to your GitHub Releases or README
   ```

**Resume execution after the user confirms keys are set up.**

---

### Task 10: Rewrite user-facing documentation

**Files:**
- Rewrite: `docs/docs/prerequisites.md`
- Rewrite: `docs/docs/installation.md`
- Modify: `docs/docs/configuration.md`
- Rewrite: `docs/docs/upgrading.md`
- Modify: `docs/docs/overview.md`
- Create: `docs/docs/uninstalling.md`
- Create: `docs/docs/nginx-setup.md`

**Step 1: Rewrite `prerequisites.md`**

Replace Docker requirements with:
- Supported OS: Fedora, RHEL, Debian, Ubuntu, Raspberry Pi OS (arm64)
- nginx (for TLS termination)
- TLS certificates (self-signed via mkcert, or Let's Encrypt)
- Network: ports 443 (nginx), 8080 (sensor-hub, localhost only)
- For temperature sensors: Raspberry Pi with Python 3.11+, DS18B20 sensor, 1-wire enabled

**Step 2: Rewrite `installation.md`**

New flow:
1. Download RPM or DEB from GitHub Releases
2. Install: `sudo dnf install ./sensor-hub-*.rpm` or `sudo dpkg -i sensor-hub_*.deb`
3. Configure `/etc/sensor-hub/application.properties` and `smtp.properties`
4. Set initial admin: edit `/etc/sensor-hub/environment`, uncomment `SENSOR_HUB_INITIAL_ADMIN`
5. Start: `sudo systemctl start sensor-hub`
6. Set up nginx (link to nginx-setup.md)
7. Verify: `curl -k https://localhost/api/health`
8. Set up OAuth (if email alerts desired)

**Step 3: Update `configuration.md`**

- Update all file paths from `configuration/` to `/etc/sensor-hub/`
- Add documentation for `--config-dir` and `--log-file` CLI flags
- Document `/etc/sensor-hub/environment` for env vars
- Update `database.path` default to `/var/lib/sensor-hub/sensor_hub.db`

**Step 4: Rewrite `upgrading.md`**

New flow:
1. Backup: `cp /var/lib/sensor-hub/sensor_hub.db ~/sensor-hub-backup.db`
2. Download new package from GitHub Releases
3. Install: `sudo dnf upgrade ./sensor-hub-*.rpm` or `sudo dpkg -i sensor-hub_*.deb`
4. Service restarts automatically (postinstall scriptlet)
5. Database migrations run automatically on startup
6. Config files preserved (noreplace)
7. Verify: `sudo systemctl status sensor-hub`

**Step 5: Create `uninstalling.md`**

Cover:
1. Stop and uninstall: `sudo dnf remove sensor-hub` or `sudo apt remove sensor-hub`
2. Note: data and config preserved by default
3. Full cleanup: `sudo rm -rf /var/lib/sensor-hub /var/log/sensor-hub /etc/sensor-hub`
4. Optional: remove user: `sudo userdel sensor-hub && sudo groupdel sensor-hub`

**Step 6: Create `nginx-setup.md`**

Cover:
1. Install nginx
2. Copy example config: `sudo cp /etc/sensor-hub/nginx.conf.example /etc/nginx/conf.d/sensor-hub.conf`
3. Edit certificate paths
4. Obtain certs (mkcert for dev, certbot for production)
5. Test config: `sudo nginx -t`
6. Enable and start: `sudo systemctl enable --now nginx`

**Step 7: Update `overview.md`**

- Replace Docker container references with "single binary" architecture
- Mention RPM/DEB packaging
- Update architecture diagram description

---

### Task 11: Create developer documentation

**Files:**
- Create: `docs/docs/development/building-from-source.md`
- Create: `docs/docs/development/docker-dev-environment.md`
- Create: `docs/docs/development/releasing.md`
- Create: `docs/docs/development/testing.md`

**Step 1: Create directory**

```bash
mkdir -p docs/docs/development
```

**Step 2: Create `building-from-source.md`**

Cover:
- Prerequisites: Go 1.25+, Node 25+, npm
- Clone repo
- Build: `cd sensor_hub && ./scripts/build.sh`
- Run locally: `./sensor-hub --config-dir=configuration`
- Dev mode without UI build: binary serves stub index.html

**Step 3: Create `docker-dev-environment.md`**

Cover:
- `cd sensor_hub && docker compose -f docker_tests/docker-compose.yml up`
- Services: sensor-hub (Air hot-reload + Delve), sensor-hub-ui (Vite HMR), mock sensors
- Ports: 8080 (API), 3000 (UI), 2345 (Delve), 5001-5002 (mock sensors)
- VSCode debug config for remote debugging

**Step 4: Create `releasing.md`**

Cover:
- Versioning: git tags (`v1.0.0`)
- Release process: `git tag v1.0.0 && git push origin v1.0.0`
- GitHub Action pipeline: what happens automatically
- GPG key setup (reference Task 9 instructions)
- Local package builds: `./scripts/build-packages.sh`
- goreleaser configuration overview

**Step 5: Create `testing.md`**

Cover:
- Running tests: `cd sensor_hub && go test ./...`
- Test structure: `_test.go` files alongside source
- Mocking patterns: sqlmock, interface-based mocks
- Test packages: list the 11 test packages

---

### Task 12: Update READMEs

**Files:**
- Modify: `README.md` (repo root)
- Modify: `sensor_hub/README.md`

**Step 1: Update root `README.md`**

- Quick start: primary path is package install (`dnf install` / `dpkg -i`)
- Secondary: link to "Building from Source" dev docs
- Remove Docker-specific quick start (or move to a "Docker" subsection)
- Keep architecture overview, API reference, feature list
- Add "Releases" badge/link

**Step 2: Update `sensor_hub/README.md`**

- Focus on developer-oriented content
- Remove deployment instructions (those live in docs/ now)
- Keep: project structure, build commands, test commands
- Link to docs/ for deployment/configuration/nginx

---

### Task 13: Final verification

**Files:** None (verification only)

**Step 1: Build**

```bash
cd sensor_hub && go build ./...
```
Expected: PASS

**Step 2: Test**

```bash
cd sensor_hub && go test ./...
```
Expected: All 11 packages PASS

**Step 3: Local package build (if goreleaser is installed)**

```bash
cd repo_root && goreleaser check
```
Expected: config is valid

```bash
./scripts/build-packages.sh
```
Expected: RPM and DEB files in `dist/`

**Step 4: Verify documentation links**

Spot-check that internal doc links point to correct files.

