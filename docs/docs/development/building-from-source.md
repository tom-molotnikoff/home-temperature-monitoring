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

## Building Packages

To build installable RPM or DEB packages locally, use the package build script.
This uses [nfpm](https://nfpm.goreleaser.com/) to produce packages identical in
structure to the CI-built releases (minus GPG signing).

### Prerequisites

| Tool | Required For     | Install                                                       |
|------|------------------|---------------------------------------------------------------|
| Go   | All builds       | [go.dev/dl](https://go.dev/dl/)                               |
| nfpm | All builds       | `go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest`    |
| npm  | Server builds    | [nodejs.org](https://nodejs.org/)                              |

### Usage

```bash
scripts/build-packages.sh <target> [options]
```

**Targets:**

| Target   | Description                                  |
|----------|----------------------------------------------|
| `cli`    | CLI-only package (no UI, no server config)   |
| `server` | Full server package (includes React UI)      |
| `all`    | Build both packages                          |

**Options:**

| Option             | Default            | Description                    |
|--------------------|--------------------|--------------------------------|
| `--arch <arch>`    | Host architecture  | `amd64` or `arm64`            |
| `--format <fmt>`   | Host-appropriate   | `rpm` or `deb`                |
| `--version <ver>`  | Auto from git tag  | Package version                |
| `--output-dir <d>` | `dist/`            | Output directory               |

### Examples

```bash
# Build CLI RPM for your machine (auto-detects everything)
scripts/build-packages.sh cli

# Build server DEB for arm64 with a specific version
scripts/build-packages.sh server --arch arm64 --format deb --version 1.2.0

# Build both packages
scripts/build-packages.sh all
```

### Version Auto-Detection

When `--version` is not specified, the script reads the latest git tag, bumps
the patch number, and appends `~dev1`. For example, if the latest tag is
`v1.1.1`, the generated version is `1.1.2~dev1`. The tilde suffix ensures the
dev version sorts above the current release but below the next release in both
RPM and DEB package managers.

### Output

Packages are written to `dist/` (or the directory specified by `--output-dir`):

```
dist/sensor-hub-cli-1.1.2~dev1-1.x86_64.rpm
dist/sensor-hub-1.1.2~dev1-1.x86_64.rpm
```

Install locally with:

```bash
# Fedora / RHEL
sudo dnf install dist/sensor-hub-cli-*.rpm

# Debian / Ubuntu
sudo apt install ./dist/sensor-hub-cli_*.deb
```
