#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VERSION="${1:?usage: render-bridge-app-formula.sh <version> <repo-full-name> <sha256s-file> [output-file]}"
REPO="${2:?usage: render-bridge-app-formula.sh <version> <repo-full-name> <sha256s-file> [output-file]}"
SHA_FILE="${3:?usage: render-bridge-app-formula.sh <version> <repo-full-name> <sha256s-file> [output-file]}"
TEMPLATE="$ROOT_DIR/packaging/homebrew/bridge-app.rb.tmpl"
OUT="${4:-$ROOT_DIR/packaging/homebrew/bridge-app.rb}"
BASE_URL="https://github.com/$REPO/releases/download/$VERSION"

lookup_sha() {
  local artifact="$1"
  local sha
  sha="$(awk -v target="./$artifact" '$2 == target { print $1 }' "$SHA_FILE")"
  if [[ -z "$sha" ]]; then
    echo "missing checksum for $artifact in $SHA_FILE" >&2
    exit 1
  fi
  printf '%s' "$sha"
}

DARWIN_ARM64="bridge-app_${VERSION}_darwin_arm64.tar.gz"
DARWIN_AMD64="bridge-app_${VERSION}_darwin_amd64.tar.gz"
LINUX_AMD64="bridge-app_${VERSION}_linux_amd64.tar.gz"
LINUX_ARM64="bridge-app_${VERSION}_linux_arm64.tar.gz"

sed \
  -e "s|__VERSION__|$VERSION|g" \
  -e "s|__DARWIN_ARM64_URL__|$BASE_URL/$DARWIN_ARM64|g" \
  -e "s|__DARWIN_ARM64_SHA256__|$(lookup_sha "$DARWIN_ARM64")|g" \
  -e "s|__DARWIN_AMD64_URL__|$BASE_URL/$DARWIN_AMD64|g" \
  -e "s|__DARWIN_AMD64_SHA256__|$(lookup_sha "$DARWIN_AMD64")|g" \
  -e "s|__LINUX_AMD64_URL__|$BASE_URL/$LINUX_AMD64|g" \
  -e "s|__LINUX_AMD64_SHA256__|$(lookup_sha "$LINUX_AMD64")|g" \
  -e "s|__LINUX_ARM64_URL__|$BASE_URL/$LINUX_ARM64|g" \
  -e "s|__LINUX_ARM64_SHA256__|$(lookup_sha "$LINUX_ARM64")|g" \
  "$TEMPLATE" > "$OUT"

echo "Wrote $OUT"
