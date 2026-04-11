#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_DIR="$ROOT_DIR/community/bridge-agent"

rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR/cmd"
mkdir -p "$TARGET_DIR/config"
mkdir -p "$TARGET_DIR/scripts"
mkdir -p "$TARGET_DIR/.github/workflows"

cp "$ROOT_DIR/LICENSE" "$TARGET_DIR/LICENSE"
cp "$ROOT_DIR/agent/go.mod" "$TARGET_DIR/go.mod"
cp "$ROOT_DIR/agent/go.sum" "$TARGET_DIR/go.sum"
cp "$ROOT_DIR/agent/cmd/agent/main.go" "$TARGET_DIR/cmd/main.go"
cp "$ROOT_DIR/config/agent.example.yaml" "$TARGET_DIR/config/agent.example.yaml"
