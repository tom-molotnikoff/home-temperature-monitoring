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
