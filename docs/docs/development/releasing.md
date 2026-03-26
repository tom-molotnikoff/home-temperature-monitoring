# Releasing

## Versioning

Releases follow [Semantic Versioning](https://semver.org/). The version is
derived from the git tag — there is no version file to update manually.
GoReleaser injects the tag into the binary via `-X main.version={{.Version}}`.

## Release Process

Tag the commit and push:

```bash
git tag v1.0.0
git push origin v1.0.0
```

That's it. The GitHub Actions workflow (`.github/workflows/release.yml`) handles
everything else.

### What Happens Automatically

1. **Checkout** with full git history (for changelog generation)
2. **Build the React UI** — `npm ci && npm run build` in `sensor_hub/ui/sensor_hub_ui`
3. **Copy UI assets** to `sensor_hub/web/dist/`
4. **Run tests** — `go test ./...`
5. **Import the GPG key** from GitHub secrets
6. **GoReleaser** cross-compiles for linux/amd64 and linux/arm64, packages RPM
   and DEB files, signs them with GPG, and uploads everything to GitHub Releases
7. The public GPG key (`sensor-hub-gpg-public.key`) is included in the release
   assets

## Local Test Build

To build packages locally without publishing or signing:

```bash
# Build CLI-only RPM for your host architecture
./scripts/build-packages.sh cli

# Build full server package (rebuilds React UI)
./scripts/build-packages.sh server

# Build both
./scripts/build-packages.sh all
```

The script uses [nfpm](https://nfpm.goreleaser.com/) directly — no GoReleaser
required. It auto-detects your host architecture, package format (RPM on
Fedora/RHEL, DEB on Debian/Ubuntu), and generates a dev version from the latest
git tag (e.g. `v1.1.1` → `1.1.2~dev1`).

Override any default:

```bash
./scripts/build-packages.sh cli --arch arm64 --format deb --version 2.0.0~beta1
```

Packages are written to `dist/`. See
[Building from Source](building-from-source.md) for prerequisites and full
usage.

## GPG Key Management

Release packages (RPM and DEB) are signed with GPG so users can verify package
integrity before installation.

| Secret / File               | Location                           | Purpose              |
|-----------------------------|------------------------------------|----------------------|
| `GPG_PRIVATE_KEY`           | GitHub repo secret                 | Signs packages       |
| `GPG_PASSPHRASE`            | GitHub repo secret                 | Unlocks private key  |
| `sensor-hub-gpg-public.key` | Repo root and every GitHub Release | Verifies signatures  |

### Rotating the GPG Key

1. Generate a new keypair:
   ```bash
   gpg --full-generate-key   # choose RSA 4096
   ```
2. Export the private key:
   ```bash
   gpg --armor --export-secret-keys FINGERPRINT > new-key.key
   ```
3. Update the GitHub repository secret `GPG_PRIVATE_KEY` with the contents of
   `new-key.key`.
4. Update `GPG_PASSPHRASE` if the passphrase changed.
5. Export the public key:
   ```bash
   gpg --armor --export FINGERPRINT > sensor-hub-gpg-public.key
   ```
6. Commit the updated `sensor-hub-gpg-public.key` to the repo root.
7. **Delete the exported private key file** — it must not be committed:
   ```bash
   rm new-key.key
   ```

> **Note:** Old releases retain their original signatures. Only new releases
> will be signed with the new key.

## GoReleaser Config Overview

The configuration lives at `.goreleaser.yml` in the repo root.

### Builds

- **Source directory**: `sensor_hub/`
- **Targets**: `linux/amd64`, `linux/arm64`
- **CGO**: disabled
- **Ldflags**: `-s -w` (strip debug info) and `-X main.version={{.Version}}`

### NFPM (System Packages)

- **Formats**: RPM and DEB
- **Binary**: installed to `/usr/bin/sensor-hub`
- **Config files** (preserved on upgrade):
  - `/etc/sensor-hub/application.properties`
  - `/etc/sensor-hub/database.properties`
  - `/etc/sensor-hub/smtp.properties`
  - `/etc/sensor-hub/environment` (mode 0640, owned by `root:sensor-hub`)
- **Systemd service**: `/usr/lib/systemd/system/sensor-hub.service`
- **Logrotate config**: `/etc/logrotate.d/sensor-hub`
- **Lifecycle scripts**: pre/post install and remove scripts in `packaging/scripts/`

### Signing

All artifacts are signed using the GPG fingerprint from the `$GPG_FINGERPRINT`
environment variable (set by the CI workflow after importing the key).
