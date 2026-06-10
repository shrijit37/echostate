#!/usr/bin/env bash
# Cross‑compile the EchoState Go daemon for all supported platforms.
# Usage: ./scripts/build-daemon.sh [version] [out-dir]
set -euo pipefail

DAEMON_DIR="daemon"
OUT_DIR="${2:-bin}"
NAME="echostate-daemon"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT}"

mkdir -p "$OUT_DIR"

build() {
    local goos="$1" goarch="$2"
    local ext=""
    [ "$goos" = "windows" ] && ext=".exe"
    local output="${OUT_DIR}/${NAME}-${goos}-${goarch}${ext}"
    echo "  ${output}"
    cd "$DAEMON_DIR"
    GOOS="$goos" GOARCH="$goarch" go build -ldflags="$LDFLAGS" -o "../${output}" .
    cd - > /dev/null
}

echo "Building EchoState daemon ${VERSION}"
build linux   amd64
build linux   arm64
build darwin  amd64
build darwin  arm64
build windows amd64

echo ""
echo "=== Build complete ==="
ls -lh "$OUT_DIR"/"${NAME}"-*
