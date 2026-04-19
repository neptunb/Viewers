#!/usr/bin/env bash
#
# Cross-compile the Go launcher for macOS (arm64+amd64), Linux (amd64) and
# Windows (amd64), then stage a distributable folder with an empty study/ dir.
#
# Assumes standalone/scripts/build-viewer.sh has been run first so that
# standalone/launcher/web/ holds the OHIF dist.
#
# Usage:
#   standalone/scripts/package.sh
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STANDALONE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
LAUNCHER_DIR="${STANDALONE_DIR}/launcher"
DIST_DIR="${STANDALONE_DIR}/dist"
EMBED_DIR="${LAUNCHER_DIR}/internal/webassets/web"

if [ ! -f "${EMBED_DIR}/index.html" ]; then
  echo "error: ${EMBED_DIR}/index.html not found. Run scripts/build-viewer.sh first." >&2
  exit 1
fi

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}/study"

cd "${LAUNCHER_DIR}"

echo "[standalone] go build macOS arm64..."
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "-s -w" \
  -o "${DIST_DIR}/macos_view_arm64" ./cmd/viewer

echo "[standalone] go build macOS amd64..."
GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "-s -w" \
  -o "${DIST_DIR}/macos_view_amd64" ./cmd/viewer

if command -v lipo >/dev/null 2>&1; then
  echo "[standalone] creating universal macOS binary..."
  lipo -create -output "${DIST_DIR}/macos_view" \
    "${DIST_DIR}/macos_view_arm64" "${DIST_DIR}/macos_view_amd64"
  rm -f "${DIST_DIR}/macos_view_arm64" "${DIST_DIR}/macos_view_amd64"
else
  mv "${DIST_DIR}/macos_view_arm64" "${DIST_DIR}/macos_view"
  rm -f "${DIST_DIR}/macos_view_amd64"
fi

echo "[standalone] go build Linux amd64..."
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "-s -w" \
  -o "${DIST_DIR}/linux_view" ./cmd/viewer

echo "[standalone] go build Windows amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "-s -w" \
  -o "${DIST_DIR}/windows_view.exe" ./cmd/viewer

cat <<EOF > "${DIST_DIR}/study/README.txt"
Drop DICOM files (or subfolders) here.
Launching macos_view / linux_view / windows_view.exe from this folder
will open the OHIF viewer automatically.
EOF

echo "[standalone] package ready: ${DIST_DIR}"
ls -lh "${DIST_DIR}"
