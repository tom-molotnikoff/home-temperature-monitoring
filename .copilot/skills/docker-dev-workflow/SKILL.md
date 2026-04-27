---
name: docker-dev-workflow
description: >
  Use when making code changes to this repository. Covers the Docker-based dev
  environment, hot-reload workflows, verification with playwright-cli, and
  common pitfalls. Read this skill BEFORE editing Go or UI code.
---

# Docker Dev Workflow

This repository uses a Docker Compose stack (`docker_tests`) for local
development. The user runs it themselves — you MUST NOT start or stop the
containers. Your job is to edit files on the host and let HMR/AIR pick up the
changes inside the running containers.

## Discovery — Always Do This First

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

If `docker_tests-sensor-hub-1` and `docker_tests-sensor-hub-ui-1` are running,
the dev stack is live. If not, ask the user to start it.

## Architecture

| Service | Container | Host Port | Tech | Hot-Reload |
|---------|-----------|-----------|------|------------|
| Go API | `docker_tests-sensor-hub-1` | `8080` | Go 1.25 + AIR + Delve | AIR watches `.go` files, rebuilds in ~10-15 s |
| React UI | `docker_tests-sensor-hub-ui-1` | `3000` | Vite + React 19 + MUI v7 | Vite HMR via polling (250 ms interval) |
| Mock sensors | `docker_tests-mock-sensor-*` | `5001`, `5002` | Python Flask | N/A |

Source directories are bind-mounted into containers, so host edits propagate
automatically.

## Hot-Reload Rules

### Go backend (AIR)

- AIR detects `.go` file changes and rebuilds automatically.
- **Expect 10-15 seconds** before the API is back. Do not assume your change
  failed if a curl immediately after editing returns an error.
- Verify readiness: `curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/health`
  (returns `200` when ready).
- AIR runs Delve (port `2345`) — the binary is a debug build.

### React UI (Vite HMR)

- Vite polls for changes every 250 ms. Most edits trigger HMR instantly.
- **If HMR does not fire**, run `touch <file>` on the edited file to force it.
  Always `touch` after using the `edit` tool to be safe.
- The dev server runs inside the container on port `5173`, mapped to host
  `localhost:3000`.

## Verifying Changes — playwright-cli

`playwright-cli` is installed globally and is the preferred way to verify UI
changes visually.

```bash
# Open the app
playwright-cli open http://localhost:3000

# Take a snapshot to see element refs
playwright-cli snapshot

# Click, type, navigate
playwright-cli click <ref>
playwright-cli fill <ref> "text"
playwright-cli goto http://localhost:3000/sensors

# Screenshot for visual verification
playwright-cli screenshot
```

Use `playwright-cli` to confirm empty states, error messages, form behavior, and
layout changes instead of relying solely on build output.

## API Verification

```bash
# Health check
curl -s http://localhost:8080/api/health

# List sensors (requires auth token)
curl -s http://localhost:8080/api/sensors/ -H "Authorization: Bearer <token>"

# Quick smoke test (no auth needed)
curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/health
```

Default admin credentials for `docker_tests`: `admin` / `adminpassword`
(the user typically overrides `SENSOR_HUB_INITIAL_ADMIN` in docker-compose).

The user may have already changed the admin password or set up a different user, so if authentication fails, ask them to provide the current credentials or check the `SENSOR_HUB_INITIAL_ADMIN` value in the `docker_tests/docker-compose.yml` file.

## Build & Lint Commands (Host)

These run **on the host**, not inside containers. Use them for pre-commit
verification.

```bash
# TypeScript type-check
cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit

# ESLint
cd sensor_hub/ui/sensor_hub_ui && npx eslint .

# Vite production build
cd sensor_hub/ui/sensor_hub_ui && npm run build

# Go build (all packages)
cd sensor_hub && go build ./...

# Go tests
cd sensor_hub && go test ./...
```

## Common Pitfalls

| Pitfall | What happens | Fix |
|---------|-------------|-----|
| Editing Go, testing immediately | 503 / connection refused | Wait 10-15 s for AIR rebuild |
| UI edit not reflecting | Stale page | `touch` the file; check browser console for HMR |
| Using `sudo` | Permission denied or policy violation | Never use sudo — ask the user to install packages |
| Running `npm run dev` on host | Port conflict with container | The container already runs the dev server |
| Running `go run .` on host | Mismatched env / DB path | The container has the correct env and volume mounts |
| Forgetting `?? []` on Go API responses | `null` in JSON crashes `.map()` | Go nil slices serialize as `null`; always guard in TS |

## Database

SQLite, stored in a Docker volume (`sensor-hub-data:/app/data`). Foreign keys
are enabled via `_pragma=foreign_keys(1)`. Go nil slices serialize as JSON
`null` — always add defensive `?? []` guards on the TypeScript side.

## File Change → Verification Workflow

```
1. Edit files on host (edit tool / vim / IDE)
2. touch edited UI files (ensures Vite HMR fires)
3. Wait for AIR rebuild if Go files changed (~15 s)
4. Verify with playwright-cli / curl / browser
5. Run tsc --noEmit + go build ./... before committing
```
