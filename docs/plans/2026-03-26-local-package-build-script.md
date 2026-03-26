# Local Package Build Script Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Replace the goreleaser-dependent `scripts/build-packages.sh` with a flexible script that builds individual packages (server or CLI-only) using `go build` + `nfpm`, with configurable architecture and format.

**Architecture:** The script uses nfpm directly with two YAML configs (`packaging/nfpm-server.yaml` and `packaging/nfpm-cli.yaml`) that mirror the goreleaser nfpm sections. It auto-detects version from the latest git tag (bump patch, append `~dev1`), host architecture, and host package format. The goreleaser config (`.goreleaser.yml`) remains for CI — only the local build script changes.

**Tech Stack:** Bash, nfpm (Go tool), Go cross-compilation via GOOS/GOARCH

---

### Task 1: Create nfpm config for CLI package

**Files:**
- Create: `packaging/nfpm-cli.yaml`

The config is a standalone nfpm YAML extracted from the `.goreleaser.yml` `sensor-hub-cli` nfpm section. It uses environment variable substitution for `version`, `arch`, and `bindir` paths.

```yaml
name: sensor-hub-cli
arch: ${ARCH}
platform: linux
version: ${VERSION}
maintainer: Tom Molotnikoff
vendor: Home Temperature Monitoring
homepage: https://github.com/tom-molotnikoff/home-temperature-monitoring
description: Sensor Hub CLI client for interacting with a remote Sensor Hub instance
license: MIT

contents:
  - src: ${BINARY_PATH}
    dst: /usr/bin/sensor-hub

  - src: packaging/completions/sensor-hub.bash
    dst: /usr/share/bash-completion/completions/sensor-hub

  - src: packaging/completions/_sensor-hub
    dst: /usr/share/zsh/site-functions/_sensor-hub

  - src: packaging/completions/sensor-hub.fish
    dst: /usr/share/fish/vendor_completions.d/sensor-hub.fish

conflicts:
  - sensor-hub

rpm:
  group: Applications/System
```

### Task 2: Create nfpm config for server package

**Files:**
- Create: `packaging/nfpm-server.yaml`

Same approach, but includes all the server-specific contents: systemd service, config files, logrotate, nginx example, and lifecycle scripts. Mirrors the `sensor-hub` nfpm section from `.goreleaser.yml`.

### Task 3: Rewrite `scripts/build-packages.sh`

**Files:**
- Modify: `scripts/build-packages.sh`

**Interface:**
```bash
scripts/build-packages.sh <target> [options]

Targets:
  cli       Build CLI-only package
  server    Build full server package (includes React UI)
  all       Build both packages

Options:
  --arch <arch>       Target architecture: amd64, arm64 (default: host)
  --format <fmt>      Package format: rpm, deb (default: host-appropriate)
  --version <ver>     Package version (default: auto from git tag)
  --output-dir <dir>  Output directory (default: dist/)
  -h, --help          Show help
```

**Logic flow:**
1. Parse arguments and set defaults (auto-detect arch, format, version)
2. Check prerequisites: `go`, `nfpm` (with install hint), `npm` (only for server)
3. Generate shell completions (build temp binary, generate, delete)
4. For server target: build React UI (`npm ci && npm run build`, copy to `web/dist/`)
5. Cross-compile Go binary: `CGO_ENABLED=0 GOOS=linux GOARCH=$ARCH go build -ldflags "-s -w -X main.version=$VERSION"`
6. Run `nfpm pkg` with the appropriate YAML config
7. Report output path

**Auto-version logic:**
```bash
latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
# Strip v prefix, split into major.minor.patch, bump patch, append ~dev1
# v1.1.1 -> 1.1.2~dev1
```

**Host detection:**
```bash
# Architecture
arch=$(uname -m)  # x86_64 -> amd64, aarch64 -> arm64

# Format
if command -v dpkg &>/dev/null; then format=deb
elif command -v rpm &>/dev/null; then format=rpm
fi
```

### Task 4: Update developer documentation — releasing.md

**Files:**
- Modify: `docs/docs/development/releasing.md`

Update the "Local Test Build" section to document the new script interface. Replace the goreleaser instructions with the new usage. Keep the CI/GoReleaser sections unchanged since those still apply to production releases.

### Task 5: Update developer documentation — building-from-source.md

**Files:**
- Modify: `docs/docs/development/building-from-source.md`

Add a "Building Packages" section at the end documenting how to build installable RPM/DEB packages locally using the new script, with examples for common workflows.

### Task 6: Test the script — build CLI RPM

Run: `scripts/build-packages.sh cli`

Verify:
- Auto-detects version from git tag
- Auto-detects host arch and format
- Produces a valid RPM in `dist/`
- `rpm -qip` shows correct metadata

### Task 7: Test the script — build with overrides

Run: `scripts/build-packages.sh cli --arch arm64 --format deb --version 2.0.0~test1`

Verify:
- Produces an arm64 DEB
- Version matches override

### Task 8: Clean up temporary build artifacts

Remove `/tmp/sensor-hub-build/` from earlier manual build.
