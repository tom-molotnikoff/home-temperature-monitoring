#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if ! command -v goreleaser &>/dev/null; then
  echo "goreleaser is not installed."
  echo "Install: https://goreleaser.com/install/"
  exit 1
fi

echo "==> Building React UI..."
cd "$REPO_ROOT/sensor_hub/ui/sensor_hub_ui"
npm ci --silent
npm run build

echo "==> Copying UI dist..."
mkdir -p "$REPO_ROOT/sensor_hub/web/dist"
cp -r dist/* "$REPO_ROOT/sensor_hub/web/dist/"

echo "==> Building packages with goreleaser (snapshot, unsigned)..."
cd "$REPO_ROOT"
goreleaser release --snapshot --skip=publish --skip=sign --clean

echo ""
echo "==> Packages built successfully:"
ls -lh dist/*.rpm dist/*.deb 2>/dev/null || echo "(no packages found — check dist/)"
