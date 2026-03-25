# Embed UI in Go Binary — Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Serve the React SPA from the Go binary via `//go:embed`, collapsing two containers into one process.

**Architecture:** Build the React app (`npm run build` → `dist/`), copy the output into `sensor_hub/web/dist/`, embed it with `//go:embed`, and serve it from a Gin `NoRoute` handler with SPA fallback. All existing API routes move under an `/api` prefix group so they don't collide with SPA client-side routes. Nginx becomes a simple TLS reverse proxy forwarding everything to the Go binary.

**Tech Stack:** Go `embed` package, Gin `NoRoute` + `http.FileServer`, existing Vite build

---

### Task 1: Create the web embedding package

**Files:**
- Create: `sensor_hub/web/embed.go`
- Create: `sensor_hub/web/handler.go`
- Create: `sensor_hub/web/dist/index.html` (dev stub)
- Modify: `sensor_hub/.gitignore`

**Step 1: Create `sensor_hub/web/embed.go`**

```go
package web

import "embed"

//go:embed all:dist
var distFS embed.FS
```

**Step 2: Create `sensor_hub/web/handler.go`**

This file provides a function that registers a Gin `NoRoute` handler serving the embedded SPA.
The handler:
- Attempts to serve the requested file from `dist/`
- If the file doesn't exist, serves `dist/index.html` (SPA fallback for client-side routing)
- Sets appropriate `Cache-Control` headers (long cache for hashed assets, no-cache for `index.html`)

```go
package web

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterSPAHandler(router *gin.Engine) {
	// Strip the "dist" prefix so that dist/index.html is served at /index.html
	stripped, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to create sub filesystem for embedded UI: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(stripped))

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try to serve the exact file (JS, CSS, images, etc.)
		if f, err := stripped.Open(strings.TrimPrefix(path, "/")); err == nil {
			f.Close()
			// Hashed asset files (e.g. /assets/index-abc123.js) can be cached aggressively
			if strings.HasPrefix(path, "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: serve index.html for client-side routing
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
```

**Step 3: Create dev stub `sensor_hub/web/dist/index.html`**

A placeholder so `go build` works without a full UI build. The real build replaces this.

```html
<!DOCTYPE html>
<html><body>
<h1>Sensor Hub UI not built</h1>
<p>Run <code>./scripts/build.sh</code> to build the full application,
or start the Vite dev server separately with <code>npm run dev</code>.</p>
</body></html>
```

**Step 4: Update `sensor_hub/.gitignore`**

Add entries so build artifacts in `web/dist/` are ignored, but the dev stub is kept:

```gitignore
web/dist/*
!web/dist/index.html
```

**Step 5: Verify `go build` works**

Run: `cd sensor_hub && go build ./...`
Expected: compiles successfully (web package embeds the stub index.html)

---

### Task 2: Move all API routes under `/api` prefix

**Files:**
- Modify: `sensor_hub/api/api.go`
- Modify: `sensor_hub/api/auth_routes.go`
- Modify: `sensor_hub/api/user_routes.go`
- Modify: `sensor_hub/api/role_routes.go`
- Modify: `sensor_hub/api/temperature_routes.go`
- Modify: `sensor_hub/api/sensor_routes.go`
- Modify: `sensor_hub/api/properties_api.go`
- Modify: `sensor_hub/api/alert_routes.go`
- Modify: `sensor_hub/api/oauth_routes.go`
- Modify: `sensor_hub/api/notification_routes.go`

**Step 1: Change all `RegisterXxxRoutes` function signatures**

Each `RegisterXxxRoutes(router *gin.Engine)` → `RegisterXxxRoutes(router gin.IRouter)`.

The `gin.IRouter` interface includes `.Group()`, `.GET()`, `.POST()`, etc. — everything
these functions currently use. This lets us pass either the root `*gin.Engine` or a
`*gin.RouterGroup`.

Verify each file uses only `gin.IRouter`-compatible methods on the router parameter
(not engine-specific methods like `.Run()` or `.Use()` on the engine level).

**Step 2: Create `/api` group in `api.go`**

In `InitialiseAndListen()`:
1. Create `apiGroup := router.Group("/api")`
2. Move the CSRF middleware to the api group: `apiGroup.Use(middleware.CSRFMiddleware())`
3. Move the health endpoint: `apiGroup.GET("/health", ...)`
4. Pass `apiGroup` instead of `router` to all `RegisterXxxRoutes` calls

The result:
- `/api/health` — health check
- `/api/auth/login` — auth routes
- `/api/sensors/` — sensor routes
- etc.

**Step 3: Integrate the SPA handler**

Add `web.RegisterSPAHandler(router)` AFTER all API route registration. This sets
up the `NoRoute` handler on the root router, catching any path not matched by `/api/*`.

```go
import "example/sensorHub/web"

// ... after all RegisterXxxRoutes(apiGroup) calls:
web.RegisterSPAHandler(router)
```

**Step 4: Make CORS conditional**

Currently CORS is always enabled. When the Go binary serves the UI itself, the browser
considers it same-origin — CORS headers are unnecessary.

Change CORS setup to only activate when `SENSOR_HUB_ALLOWED_ORIGIN` is explicitly set:
```go
allowedOrigin := os.Getenv("SENSOR_HUB_ALLOWED_ORIGIN")
if allowedOrigin != "" {
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{allowedOrigin},
        // ... rest unchanged
    }))
}
```

This keeps CORS available for development (Vite dev server on a different port) but
skips it in production where everything is same-origin.

**Step 5: Run tests**

Run: `cd sensor_hub && go test ./api/... ./service/...`
Expected: all tests pass.

Some API tests may mock the router; verify they don't hardcode route paths without
the `/api` prefix. If tests call endpoints directly, they need the new paths.

---

### Task 3: Update API test route paths

**Files:**
- Modify: all test files in `sensor_hub/api/` that call route paths
- Modify: `sensor_hub/service/` test files if they reference route paths

**Step 1: Search for hardcoded route paths in tests**

Search all `_test.go` files in `api/` and `service/` for `httptest.NewRequest` or
`req, _ := http.NewRequest` calls that use paths like `"/auth/login"`, `"/sensors/"`, etc.

**Step 2: Prefix all test route paths with `/api`**

Every test request path needs updating, e.g.:
- `"/auth/login"` → `"/api/auth/login"`
- `"/health"` → `"/api/health"`
- `"/sensors/"` → `"/api/sensors/"`
- etc.

**Step 3: Run API tests**

Run: `cd sensor_hub && go test ./api/...`
Expected: all tests pass with new route paths.

---

### Task 4: Create the build script

**Files:**
- Create: `sensor_hub/scripts/build.sh`

**Step 1: Create `sensor_hub/scripts/build.sh`**

```bash
#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
UI_DIR="$PROJECT_DIR/ui/sensor_hub_ui"
WEB_DIST="$PROJECT_DIR/web/dist"

echo "=== Building Sensor Hub ==="

# Step 1: Build the React UI
echo ""
echo "--- Building UI ---"
cd "$UI_DIR"
npm ci --silent
npm run build
echo "UI build complete."

# Step 2: Copy build output to web/dist/
echo ""
echo "--- Copying UI to web/dist/ ---"
rm -rf "$WEB_DIST"
cp -r "$UI_DIR/dist" "$WEB_DIST"
echo "Copied $(find "$WEB_DIST" -type f | wc -l) files."

# Step 3: Build the Go binary
echo ""
echo "--- Building Go binary ---"
cd "$PROJECT_DIR"
go build -o sensor-hub .
echo "Binary: $PROJECT_DIR/sensor-hub ($(du -h sensor-hub | cut -f1))"

echo ""
echo "=== Build complete ==="
```

**Step 2: Make executable**

Run: `chmod +x sensor_hub/scripts/build.sh`

---

### Task 5: Update nginx config for simple reverse proxy

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/nginx/default.conf`

**Step 1: Simplify the nginx config**

Nginx no longer serves static files or strips `/api/` prefix. It's now a pure TLS
reverse proxy, forwarding everything to the Go binary:

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ""      close;
}

server {
    listen 80;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name _;

    ssl_certificate /etc/ssl/certs/home.sensor-hub.pem;
    ssl_certificate_key /etc/ssl/private/home.sensor-hub-key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Key changes:
- Removed separate `/api/` and `/ws` locations — all traffic goes to Go
- `proxy_pass http://` (plain HTTP) — Go binary no longer needs TLS since nginx handles it
- WebSocket upgrade headers are on the single `location /` block

---

### Task 6: Update Docker Compose to single service

**Files:**
- Modify: `sensor_hub/docker/docker-compose.yml`
- Modify: `sensor_hub/docker/sensor-hub.dockerfile`

**Step 1: Update docker-compose.yml**

Remove the `sensor-hub-ui` service entirely. The Go binary serves everything:

```yaml
services:
  sensor-hub:
    build:
      context: ..
      dockerfile: docker/sensor-hub.dockerfile
    ports:
      - "8080:8080"
    volumes:
      - sensor-hub-data:/app/data
    environment:
      SENSOR_HUB_ALLOWED_ORIGIN: ""  # same-origin, no CORS needed
volumes:
  sensor-hub-data:
```

Note: TLS certificate mounts are removed from the Go container since nginx handles TLS
on the host. If the user wants in-container TLS, they can re-add the cert volume mounts
and set `TLS_CERT_FILE`/`TLS_KEY_FILE`.

**Step 2: Update sensor-hub.dockerfile for multi-stage build**

The Dockerfile now builds the UI and the Go binary:

```dockerfile
# Stage 1: Build the React UI
FROM node:25-alpine AS ui-build
WORKDIR /ui
COPY ui/sensor_hub_ui/package.json ui/sensor_hub_ui/package-lock.json ./
RUN npm ci --silent
COPY ui/sensor_hub_ui/ ./
RUN npm run build

# Stage 2: Build the Go binary with embedded UI
FROM golang:1.25-alpine AS go-build
WORKDIR /app
COPY sensor_hub/ ./
COPY --from=ui-build /ui/dist ./web/dist/
RUN go mod download
RUN go build -o sensor-hub .

# Stage 3: Minimal runtime
FROM alpine:latest
WORKDIR /app
COPY --from=go-build /app/sensor-hub .
COPY --from=go-build /app/configuration/ ./configuration/
CMD ["./sensor-hub"]
```

---

### Task 7: Update docker_tests compose

**Files:**
- Modify: `sensor_hub/docker_tests/docker-compose.yml`
- Modify: `sensor_hub/docker_tests/sensor-hub.dockerfile`

Apply analogous changes as Task 6:
- Remove `sensor-hub-ui` service from docker_tests compose
- Update the test Dockerfile to include the UI build (or use the dev stub for tests)
- Keep mock sensor services unchanged

For the test Dockerfile, since it uses `air` for hot-reloading, the dev stub is
sufficient — no need to build the full UI for backend tests.

---

### Task 8: Build and verify

**Step 1: Run the build script**

Run: `cd sensor_hub && ./scripts/build.sh`
Expected: UI builds, binary compiles with embedded UI

**Step 2: Run all tests**

Run: `cd sensor_hub && go test ./...`
Expected: all tests pass

**Step 3: Smoke test the binary**

Run the binary and verify:
1. `curl http://localhost:8080/api/health` → `{"status":"ok"}`
2. `curl http://localhost:8080/` → returns `index.html` (the React SPA)
3. `curl http://localhost:8080/sensors` → returns `index.html` (SPA fallback)
4. `curl http://localhost:8080/api/sensors/` → returns JSON (API response or 401)

---

### Task 9: Update documentation

**Files:**
- Modify: `README.md`
- Modify: `sensor_hub/README.md`
- Modify: `docs/docs/overview.md`
- Modify: `docs/docs/installation.md`
- Modify: `docs/docs/configuration.md`
- Modify: `docs/docs/prerequisites.md`

Update to reflect:
- Single binary serves both API and UI
- Build process: `./scripts/build.sh` produces the binary
- Nginx is a simple TLS reverse proxy (no static file serving, no `/api` stripping)
- `SENSOR_HUB_ALLOWED_ORIGIN` is optional (only needed for separate dev server)
- Docker Compose runs a single container
- No separate UI container needed
