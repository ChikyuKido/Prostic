#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"
DIST_DIR="$ROOT_DIR/internal/embed/dist"
BINARY_NAME="${1:-prostic}"
GO_CACHE_DIR="${GOCACHE:-$ROOT_DIR/.cache/go-build}"

mkdir -p "$GO_CACHE_DIR"

echo "[*] Building web UI..."
(
  cd "$WEB_DIR"
  bun run build
)

echo "[*] Preparing embedded dist directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
cp -r "$WEB_DIR/dist/." "$DIST_DIR/"

echo "[*] Compressing web assets with gzip..."
"$ROOT_DIR/scripts/gzip-dist.sh" "$DIST_DIR"

echo "[*] Building Go binary with embedded assets..."
(
  cd "$ROOT_DIR"
  GOCACHE="$GO_CACHE_DIR" GOOS=linux GOARCH=amd64 go build -o "$BINARY_NAME" ./cmd
)

echo "[✓] Built $BINARY_NAME"
