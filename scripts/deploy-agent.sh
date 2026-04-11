#!/usr/bin/env bash
# deploy-agent.sh — build and deploy bridge-agent to a remote device over SSH.
#
# Usage:
#   bash scripts/deploy-agent.sh <ssh-target>
#
# Examples:
#   bash scripts/deploy-agent.sh pi@100.x.x.x
#   bash scripts/deploy-agent.sh pi@raspberry-pi.tail1234.ts.net
#
# What it does:
#   1. Cross-compiles bridge-agent for the target OS/arch (auto-detected).
#   2. Copies binary + installer + example config to the device.
#   3. SSHs in and runs install-agent.sh interactively.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SSH_TARGET="${1:-}"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

info()    { printf '\033[0;34m[deploy]\033[0m %s\n' "$*"; }
success() { printf '\033[0;32m[deploy]\033[0m %s\n' "$*"; }
err()     { printf '\033[0;31m[deploy]\033[0m %s\n' "$*" >&2; }

if [[ -z "$SSH_TARGET" ]]; then
  err "Usage: bash scripts/deploy-agent.sh <ssh-target>"
  err "Example: bash scripts/deploy-agent.sh pi@100.x.x.x"
  exit 1
fi

# ---------------------------------------------------------------------------
# Detect target OS and arch
# ---------------------------------------------------------------------------

info "Detecting target platform for $SSH_TARGET ..."
TARGET_OS="$(ssh "$SSH_TARGET" 'uname -s' 2>/dev/null | tr '[:upper:]' '[:lower:]')"
TARGET_ARCH_RAW="$(ssh "$SSH_TARGET" 'uname -m' 2>/dev/null)"

case "$TARGET_ARCH_RAW" in
  x86_64)           TARGET_ARCH="amd64" ;;
  aarch64|arm64)    TARGET_ARCH="arm64" ;;
  armv7l|armv6l)    TARGET_ARCH="arm" ;;
  *)
    err "Unsupported architecture: $TARGET_ARCH_RAW"
    exit 1
    ;;
esac

info "Target: $TARGET_OS/$TARGET_ARCH"

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

BUILD_OUT="$ROOT_DIR/agent/bridge-agent-$TARGET_OS-$TARGET_ARCH"

info "Building bridge-agent for $TARGET_OS/$TARGET_ARCH ..."
(
  cd "$ROOT_DIR/agent"
  GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -o "$BUILD_OUT" ./cmd/agent
)
success "Build complete: $BUILD_OUT"

# ---------------------------------------------------------------------------
# Copy to device
# ---------------------------------------------------------------------------

REMOTE_DIR="~/bridge-agent"

info "Copying files to $SSH_TARGET:$REMOTE_DIR ..."
ssh "$SSH_TARGET" "mkdir -p $REMOTE_DIR"
scp "$BUILD_OUT" "$SSH_TARGET:$REMOTE_DIR/bridge-agent"
scp "$ROOT_DIR/scripts/install-agent.sh" "$SSH_TARGET:$REMOTE_DIR/install-agent.sh"
scp "$ROOT_DIR/config/agent.example.yaml" "$SSH_TARGET:$REMOTE_DIR/agent.yaml.example"
success "Files copied."

# ---------------------------------------------------------------------------
# Run installer on device
# ---------------------------------------------------------------------------

info "Running installer on $SSH_TARGET ..."
echo
ssh -t "$SSH_TARGET" "cd $REMOTE_DIR && bash install-agent.sh"
