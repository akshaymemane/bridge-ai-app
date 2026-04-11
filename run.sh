#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE="${1:-all}"
AGENT_CONFIG="${AGENT_CONFIG:-$ROOT_DIR/agent/agent.yaml}"

usage() {
  cat <<EOF
Usage: ./run.sh [all|gateway|agent|ui|help]

Modes:
  all      Run gateway, agent, and frontend together
  gateway  Run only the Go gateway
  agent    Run only the bridge agent
  ui       Run only the frontend dev server
  help     Show this help

Optional env:
  AGENT_CONFIG=/absolute/path/to/agent.yaml
EOF
}

require_file() {
  local file_path="$1"
  local label="$2"

  if [[ ! -f "$file_path" ]]; then
    echo "Missing $label: $file_path" >&2
    exit 1
  fi
}

run_gateway() {
  cd "$ROOT_DIR/gateway"
  go run ./cmd/gateway
}

run_agent() {
  require_file "$AGENT_CONFIG" "agent config"
  cd "$ROOT_DIR/agent"
  go run ./cmd/agent -config "$AGENT_CONFIG"
}

run_ui() {
  cd "$ROOT_DIR/frontend"
  npm run dev
}

run_all() {
  require_file "$AGENT_CONFIG" "agent config"

  local gateway_pid=""
  local agent_pid=""
  local ui_pid=""

  cleanup() {
    local exit_code=$?
    trap - EXIT INT TERM

    for pid in "$ui_pid" "$agent_pid" "$gateway_pid"; do
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
    cd "$ROOT_DIR/agent"
    exec go run ./cmd/agent -config "$AGENT_CONFIG"
  ) &
  agent_pid=$!
  echo "Agent started (pid $agent_pid)"

  (
    cd "$ROOT_DIR/frontend"
    exec npm run dev
  ) &
  ui_pid=$!
  echo "UI started (pid $ui_pid)"

  echo "BridgeAIChat is starting. Press Ctrl+C to stop all services."
  wait
}

case "$MODE" in
  all)
    run_all
    ;;
  gateway)
    run_gateway
    ;;
  agent)
    run_agent
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
