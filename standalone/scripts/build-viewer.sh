#!/usr/bin/env bash
#
# Build the OHIF viewer with the standalone config and copy the dist into
# standalone/launcher/web/ so `go build` can embed it via //go:embed.
#
# Usage:
#   standalone/scripts/build-viewer.sh
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STANDALONE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${STANDALONE_DIR}/.." && pwd)"

CONFIG_SRC="${STANDALONE_DIR}/viewer-config/standalone.js"
CONFIG_DST_DIR="${REPO_ROOT}/platform/app/public/config"
CONFIG_DST="${CONFIG_DST_DIR}/standalone.js"

VIEWER_DIST="${REPO_ROOT}/platform/app/dist"
EMBED_DIR="${STANDALONE_DIR}/launcher/internal/webassets/web"

echo "[standalone] repo:      ${REPO_ROOT}"
echo "[standalone] config:    ${CONFIG_SRC}"
echo "[standalone] embed dir: ${EMBED_DIR}"

if [ ! -f "${CONFIG_SRC}" ]; then
  echo "error: config not found: ${CONFIG_SRC}" >&2
  exit 1
fi

mkdir -p "${CONFIG_DST_DIR}"
cp "${CONFIG_SRC}" "${CONFIG_DST}"
echo "[standalone] copied standalone.js -> ${CONFIG_DST}"

cd "${REPO_ROOT}"
if [ ! -d "node_modules" ]; then
  echo "[standalone] installing yarn dependencies..."
  yarn install --frozen-lockfile
fi

echo "[standalone] building viewer (APP_CONFIG=config/standalone.js)..."
APP_CONFIG=config/standalone.js \
PUBLIC_URL=/ \
  yarn build

if [ ! -f "${VIEWER_DIST}/index.html" ]; then
  echo "error: build did not produce ${VIEWER_DIST}/index.html" >&2
  exit 1
fi

rm -rf "${EMBED_DIR}"
mkdir -p "${EMBED_DIR}"
cp -R "${VIEWER_DIST}/." "${EMBED_DIR}/"
echo "[standalone] copied viewer dist -> ${EMBED_DIR}"
echo "[standalone] done."
