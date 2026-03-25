# RPM/DEB Packaging Design

## Overview

Package the sensor-hub application as RPM and DEB packages for direct installation on Linux hosts, eliminating the need to clone the repo and run Docker in production. The binary, systemd unit, config files, and log rotation are all managed by the package.

## Architecture

### Deployment Stack (Production)

```
[nginx] --TLS--> [sensor-hub binary] --SQLite--> /var/lib/sensor-hub/sensor_hub.db
  :443              :8080 (localhost)
```

nginx is not shipped in the package — an example config is provided and documented.

### FHS File Layout

| Path | Contents | Permissions | Package Behaviour |
|------|----------|-------------|-------------------|
| `/usr/bin/sensor-hub` | Go binary (API + embedded UI) | `0755 root:root` | Replaced on upgrade |
| `/usr/lib/systemd/system/sensor-hub.service` | Systemd unit | `0644 root:root` | Replaced on upgrade |
| `/etc/sensor-hub/application.properties` | App config | `0640 root:sensor-hub` | `config(noreplace)` |
| `/etc/sensor-hub/database.properties` | DB path config | `0640 root:sensor-hub` | `config(noreplace)` |
| `/etc/sensor-hub/smtp.properties` | SMTP config | `0640 root:sensor-hub` | `config(noreplace)` |
| `/etc/sensor-hub/environment` | Env vars for systemd | `0640 root:sensor-hub` | `config(noreplace)` |
| `/etc/sensor-hub/nginx.conf.example` | Reference nginx config | `0644 root:root` | Replaced on upgrade |
| `/etc/logrotate.d/sensor-hub` | Log rotation policy | `0644 root:root` | Replaced on upgrade |
| `/var/lib/sensor-hub/` | SQLite database | `0750 sensor-hub:sensor-hub` | Created by scriptlet, preserved on uninstall |
| `/var/log/sensor-hub/` | Application logs | `0750 sensor-hub:sensor-hub` | Created by scriptlet, preserved on uninstall |

Config files are marked `config(noreplace)` — user edits survive upgrades; new defaults saved as `.rpmnew`/`.dpkg-new`.

### File Permissions Rationale

- `/etc/sensor-hub/` is `0750 root:sensor-hub` — root owns the files, the service user can read them.
- Properties files are `0640` — not world-readable (SMTP credentials, sensitive config).
- `/var/lib/sensor-hub/` is `0750 sensor-hub:sensor-hub` — only the service user can read/write the database (contains password hashes).
- `/var/log/sensor-hub/` is `0750 sensor-hub:sensor-hub` — logs may contain sensitive request data.

## Scriptlets

### Pre-install

```bash
# Create system user (idempotent)
getent group sensor-hub >/dev/null || groupadd --system sensor-hub
getent passwd sensor-hub >/dev/null || useradd --system \
  --gid sensor-hub \
  --home-dir /var/lib/sensor-hub \
  --shell /usr/sbin/nologin \
  --comment "Sensor Hub service account" \
  sensor-hub
```

### Post-install

```bash
# Ensure directories exist with correct ownership
install -d -m 0750 -o sensor-hub -g sensor-hub /var/lib/sensor-hub
install -d -m 0750 -o sensor-hub -g sensor-hub /var/log/sensor-hub

systemctl daemon-reload

# First install: enable but don't start (user must configure first)
if [ "$1" = "1" ] || [ "$1" = "configure" ]; then
  systemctl enable sensor-hub
  echo "Sensor Hub installed. Configure /etc/sensor-hub/ then run: systemctl start sensor-hub"
fi

# Upgrade: restart to pick up new binary
if [ "$1" = "2" ] || [ "$1" = "upgrade" ]; then
  systemctl restart sensor-hub
fi
```

### Pre-uninstall

```bash
if [ "$1" = "0" ] || [ "$1" = "remove" ]; then
  systemctl stop sensor-hub
  systemctl disable sensor-hub
fi
```

### Post-uninstall

```bash
systemctl daemon-reload

# Data and config are intentionally preserved.
# To fully remove: rm -rf /var/lib/sensor-hub /var/log/sensor-hub /etc/sensor-hub
```

The `sensor-hub` user is intentionally NOT removed on uninstall (standard practice to prevent UID reuse and orphaned file ownership).

## Systemd Service Unit

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

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/sensor-hub /var/log/sensor-hub
PrivateTmp=yes

# Environment
Environment=SENSOR_HUB_PRODUCTION=true
EnvironmentFile=-/etc/sensor-hub/environment

[Install]
WantedBy=multi-user.target
```

The `EnvironmentFile=-` (dash prefix = optional) lets users set runtime env vars (`SENSOR_HUB_INITIAL_ADMIN`, `TLS_CERT_FILE`, etc.) without editing the unit file.

## Logrotate Configuration

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

Rotates daily or at 50MB (whichever first). Keeps 14 days compressed. Uses `copytruncate` so the binary doesn't need a signal to reopen the file.

## Build & Release Pipeline

### Tooling

- **goreleaser** — Handles Go cross-compilation, nfpm packaging (RPM + DEB), GPG signing, and GitHub Release upload in a single tool.
- **nfpm** — Used internally by goreleaser for RPM/DEB generation from YAML config.

### Versioning

- Single source of truth: git tags (e.g., `v1.2.0`)
- goreleaser extracts version from the tag
- Version baked into binary via `-ldflags "-X main.version={{.Version}}"`
- `sensor-hub --version` prints the version

### Target Architectures

- `linux/amd64` — Development/testing (Fedora)
- `linux/arm64` — Production (Raspberry Pi 5)

### GitHub Action (`.github/workflows/release.yml`)

Triggers on tag push matching `v*`:

1. Checkout code
2. Setup Node 25 → `npm ci && npm run build` → copy dist to `web/dist/`
3. Setup Go 1.25
4. Import GPG key from GitHub secrets
5. Run `goreleaser release`
6. Packages + checksums uploaded to GitHub Releases

### GitHub Secrets Required

| Secret | Purpose |
|--------|---------|
| `GPG_PRIVATE_KEY` | ASCII-armored GPG private key for package signing |
| `GPG_PASSPHRASE` | Passphrase for the GPG key |

### Local Build Script (`scripts/build-packages.sh`)

Builds unsigned packages locally for testing:

```bash
goreleaser release --snapshot --skip=publish
# Output in dist/
```

### Release Workflow

```bash
git tag v1.0.0
git push origin v1.0.0
# → GitHub Action triggers → builds → signs → publishes to Releases page
```

## Code Changes Required

### Existing Code Modifications

1. **`main.go`** — Add `--config-dir`, `--log-file`, and `--version` CLI flags. Add `var version string` populated via ldflags.

2. **`application_properties/`** — Accept config directory as parameter instead of hardcoded `./configuration/` path.

3. **`go.mod`** — The module is currently `example/sensorHub`. This needs to change to a real module path for goreleaser to work properly (goreleaser uses `go build` with the module path).

### New Files

```
.goreleaser.yml                          # goreleaser + nfpm config
.github/workflows/release.yml           # Tag-triggered release pipeline
packaging/
  sensor-hub.service                     # Systemd unit file
  logrotate.conf                         # Logrotate config
  nginx.conf.example                     # Reference nginx config
  environment                            # Template env vars file
  defaults/
    application.properties               # Production defaults
    database.properties                   # database.path=/var/lib/sensor-hub/sensor_hub.db
    smtp.properties                      # Empty SMTP config
  scripts/
    preinstall.sh                        # Create system user
    postinstall.sh                       # Create dirs, enable service
    preremove.sh                         # Stop and disable service
    postremove.sh                        # Daemon reload
scripts/
  build-packages.sh                      # Local unsigned package build
```

## nginx

nginx is NOT a dependency of the package. The package ships `/etc/sensor-hub/nginx.conf.example` as a reference config. Documentation covers:

1. Installing nginx (`dnf install nginx` / `apt install nginx`)
2. Copying and adapting the example config
3. Obtaining TLS certificates (certbot or self-signed)
4. Enabling and starting nginx

## Upgrade Path

1. User runs `dnf upgrade sensor-hub` or `dpkg -i sensor-hub_x.y.z.deb`
2. Pre-install scriptlet is a no-op (user already exists)
3. Package manager replaces binary and unit file
4. Config files preserved (noreplace)
5. Post-install scriptlet detects upgrade (`$1 = 2` / `$1 = upgrade`) and restarts the service
6. SQLite migrations run automatically on startup (golang-migrate)

## Documentation Overhaul

The existing docs are heavily Docker-oriented and assume users clone the repo and run `docker compose up`. With RPM/DEB packaging, the docs need a restructure to separate **user-facing deployment documentation** from **developer documentation**.

### User-Facing Documentation (rewrite)

These docs are for people deploying and operating the application:

| File | Current State | Action |
|------|---------------|--------|
| `docs/docs/prerequisites.md` | Lists Docker as a requirement | Rewrite: system requirements (Go not needed — it's a binary), nginx, TLS certs, network ports |
| `docs/docs/installation.md` | Docker Compose clone-and-run workflow | Rewrite: `dnf install` / `dpkg -i` from GitHub Releases, configure `/etc/sensor-hub/`, set up nginx, start service |
| `docs/docs/configuration.md` | References `./configuration/` paths | Update: paths to `/etc/sensor-hub/`, document `/etc/sensor-hub/environment` for env vars, document `--config-dir` and `--log-file` flags |
| `docs/docs/upgrading.md` | `git pull && docker compose build` | Rewrite: `dnf upgrade` / `dpkg -i`, config preservation (noreplace), migration notes |
| `docs/docs/overview.md` | References Docker containers | Update: describe single-binary + nginx architecture, remove Docker-specific language |
| `README.md` | Quick start is build-from-source | Update: primary quick start is package install, link to dev docs for contributors |
| `sensor_hub/README.md` | Build and Docker instructions | Update: remove deployment content, keep as developer-oriented README |

Additionally, create:

| File | Purpose |
|------|---------|
| `docs/docs/uninstalling.md` | How to uninstall the package and optionally remove data/config/user |
| `docs/docs/nginx-setup.md` | Dedicated nginx + TLS setup guide (from the example config shipped in the package) |

### Developer Documentation (new section)

Create a development section in the docs for contributors and anyone building from source. This content currently lives scattered across READMEs.

| File | Purpose |
|------|---------|
| `docs/docs/development/building-from-source.md` | Prerequisites (Go 1.25, Node 25), `scripts/build.sh`, running locally, `--config-dir` for dev |
| `docs/docs/development/docker-dev-environment.md` | `docker_tests/docker-compose.yml`, Vite HMR, Air hot-reload, Delve debugging |
| `docs/docs/development/releasing.md` | Versioning (git tags), goreleaser, GPG key setup, `scripts/build-packages.sh` for local builds, GitHub Action release pipeline, how to cut a release |
| `docs/docs/development/testing.md` | Running tests (`go test ./...`), test structure, mocks |

### Documentation Task Summary

1. Rewrite `prerequisites.md` — system requirements for package-based install
2. Rewrite `installation.md` — package install + nginx setup + first-run config
3. Update `configuration.md` — FHS paths, environment file, CLI flags
4. Rewrite `upgrading.md` — package manager upgrade workflow
5. Create `uninstalling.md` — clean uninstall instructions
6. Create `nginx-setup.md` — dedicated nginx + TLS guide
7. Update `overview.md` — single-binary architecture description
8. Update `README.md` — package install as primary quick start
9. Update `sensor_hub/README.md` — developer-focused content only
10. Create `development/building-from-source.md` — build prerequisites and process
11. Create `development/docker-dev-environment.md` — Docker dev stack guide
12. Create `development/releasing.md` — release process, GPG keys, goreleaser, GitHub Actions
13. Create `development/testing.md` — test running and structure

## Security Considerations

- Database file (`/var/lib/sensor-hub/`) contains bcrypt password hashes — restricted to `sensor-hub` user only
- Config files may contain SMTP credentials — not world-readable
- `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome=yes` in systemd unit
- GPG-signed packages for integrity verification
- `PrivateTmp=yes` prevents temp file snooping
