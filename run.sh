#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-all}"

usage() {
  cat <<EOF
Usage: ./run.sh [all|gateway|ui|help]

Modes:
  all      Run gateway and frontend together
  gateway  Run only the Go gateway
  ui       Run only the frontend dev server
  help     Show this help
EOF
}

run_gateway() {
  cd "$ROOT_DIR/gateway"
  go run ./cmd/gateway
}

run_ui() {
  cd "$ROOT_DIR/frontend"
  npm run dev
}

run_all() {
  local gateway_pid=""
  local ui_pid=""

  cleanup() {
    local exit_code=$?
    trap - EXIT INT TERM

    for pid in "$ui_pid" "$gateway_pid"; do
      if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
      fi
    done

    wait || true
    exit "$exit_code"
  }

  trap cleanup EXIT INT TERM

  (
    cd "$ROOT_DIR/gateway"
    exec go run ./cmd/gateway
  ) &
  gateway_pid=$!
  echo "Gateway started (pid $gateway_pid)"

  (
    cd "$ROOT_DIR/frontend"
    exec npm run dev
  ) &
  ui_pid=$!
  echo "UI started (pid $ui_pid)"

  echo "BridgeAIChat app is starting. Press Ctrl+C to stop both services."
  wait
}

case "$MODE" in
  all)
    run_all
    ;;
  gateway)
    run_gateway
    ;;
  ui)
    run_ui
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    echo "Unknown mode: $MODE" >&2
    echo >&2
    usage >&2
    exit 1
    ;;
esac
