# Building from Source

## Prerequisites

| Tool   | Version | Purpose                        |
|--------|---------|--------------------------------|
| Go     | 1.25+   | Backend compilation            |
| Node   | 25+     | React UI build                 |
| npm    | latest  | JavaScript dependency manager  |
| git    | latest  | Source control                 |

## Clone and Build

```bash
git clone https://github.com/tom-molotnikoff/home-temperature-monitoring.git
cd home-temperature-monitoring/sensor_hub
./scripts/build.sh
```

The build script performs three steps:

1. **Install UI dependencies** — runs `npm ci` in `ui/sensor_hub_ui`
2. **Build the React UI** — runs `npm run build`, copies output to `web/dist/`
3. **Compile the Go binary** — runs `go build -o sensor-hub .`

The resulting `sensor-hub` binary embeds the UI static assets from `web/dist/`.

## Run Locally

```bash
./sensor-hub --config-dir=configuration
```

The binary serves on **port 8080** by default.

## Dev Mode (Without Full UI Build)

When working on the Go backend, you can skip the full UI build. Create a stub so
`go build` can embed the `web/dist/` directory:

```bash
mkdir -p web/dist
echo '<!doctype html><html><body>stub</body></html>' > web/dist/index.html
go build -o sensor-hub .
```

To develop the React UI with hot module replacement, run the Vite dev server
separately:

```bash
cd ui/sensor_hub_ui
npm install
npm run dev
```

The Vite dev server starts on port 5173. Point it at the Go backend by setting
the `VITE_API_BASE` and `VITE_WEBSOCKET_BASE` environment variables (see
[Docker Dev Environment](docker-dev-environment.md) for the full list).
