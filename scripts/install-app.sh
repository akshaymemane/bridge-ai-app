#!/usr/bin/env bash
# install-app.sh — install bridge-app from an extracted release directory.
#
# Usage:
#   bash install-app.sh
#
# What it does:
#   1. Copies the gateway binary and UI assets into ~/.bridge-app
#   2. Installs a bridge-gateway launcher into ~/.local/bin
#   3. Optionally installs a startup service

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_ROOT="$HOME/.bridge-app"
CURRENT_DIR="$APP_ROOT/current"
BIN_DIR="$HOME/.local/bin"
INSTALLED_BIN="$BIN_DIR/bridge-gateway"
SRC_BIN="$SCRIPT_DIR/bridge-gateway"
SRC_UI="$SCRIPT_DIR/ui"

info()    { printf '\033[0;34m[bridge-app]\033[0m %s\n' "$*"; }
success() { printf '\033[0;32m[bridge-app]\033[0m %s\n' "$*"; }
warn()    { printf '\033[0;33m[bridge-app]\033[0m %s\n' "$*"; }
err()     { printf '\033[0;31m[bridge-app]\033[0m %s\n' "$*" >&2; }

install_launchd() {
  local plist_dir="$HOME/Library/LaunchAgents"
  local plist_path="$plist_dir/com.bridge-app.gateway.plist"
  mkdir -p "$plist_dir"

  cat > "$plist_path" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.bridge-app.gateway</string>
  <key>ProgramArguments</key>
  <array>
    <string>$INSTALLED_BIN</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>$HOME/.bridge-app.log</string>
  <key>StandardErrorPath</key>
  <string>$HOME/.bridge-app.log</string>
</dict>
</plist>
EOF

  launchctl load "$plist_path" 2>/dev/null || true
  success "launchd service installed."
  info "Log file: $HOME/.bridge-app.log"
}

install_systemd() {
  local service_path="/etc/systemd/system/bridge-app-gateway.service"
  local current_user
  current_user="$(id -un)"

  if ! sudo tee "$service_path" > /dev/null <<EOF
[Unit]
Description=BridgeAIChat gateway
After=network.target

[Service]
Type=simple
User=$current_user
ExecStart=$INSTALLED_BIN
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
  then
    err "Failed to write systemd service."
    return
  fi

  sudo systemctl daemon-reload
  sudo systemctl enable bridge-app-gateway
  sudo systemctl start bridge-app-gateway
  success "systemd service installed and started."
}

info "Starting bridge-app installer"

if [[ ! -f "$SRC_BIN" ]]; then
  err "bridge-gateway binary not found in $SCRIPT_DIR"
  exit 1
fi

if [[ ! -d "$SRC_UI" ]]; then
  err "ui directory not found in $SCRIPT_DIR"
  exit 1
fi

mkdir -p "$CURRENT_DIR" "$BIN_DIR"
chmod +x "$SRC_BIN"

info "Installing files into $CURRENT_DIR"
cp "$SRC_BIN" "$CURRENT_DIR/bridge-gateway"
rm -rf "$CURRENT_DIR/ui"
cp -R "$SRC_UI" "$CURRENT_DIR/ui"

cat > "$INSTALLED_BIN" <<EOF
#!/usr/bin/env bash
set -euo pipefail
exec "$CURRENT_DIR/bridge-gateway" -ui-dist "$CURRENT_DIR/ui"
EOF
chmod +x "$INSTALLED_BIN"

success "bridge-app installed"
info "Launcher installed to: $INSTALLED_BIN"
info "If ~/.local/bin is not on PATH, run with: $INSTALLED_BIN"

read -r -p "  Install gateway as a startup service? [y/N] " install_service
if [[ "${install_service,,}" == "y" ]]; then
  case "$(uname)" in
    Darwin) install_launchd ;;
    Linux) install_systemd ;;
    *) warn "Unsupported OS for automatic service install." ;;
  esac
fi

echo
success "Installation complete."
info "Start the gateway with:"
echo "    $INSTALLED_BIN"
info "Then open:"
echo "    http://localhost:8080"
