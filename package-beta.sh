#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="${1:-beta-$(date +%Y%m%d-%H%M%S)}"
OUT_DIR="$ROOT_DIR/dist/$VERSION"

mkdir -p "$OUT_DIR"
mkdir -p "$OUT_DIR/gateway" "$OUT_DIR/docs"

echo "Packaging BridgeAIChat $VERSION"
echo "Output: $OUT_DIR"

(
  cd "$ROOT_DIR/frontend"
  npm run build
)

(
  cd "$ROOT_DIR/gateway"
  go build -o "$OUT_DIR/gateway/bridge-gateway" ./cmd/gateway
)

cp -R "$ROOT_DIR/frontend/dist" "$OUT_DIR/gateway/ui"

cat > "$OUT_DIR/gateway/run-gateway.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$DIR/bridge-gateway" -ui-dist "$DIR/ui"
EOF
chmod +x "$OUT_DIR/gateway/run-gateway.sh"

cp "$ROOT_DIR/BETA_TESTING.md" "$OUT_DIR/docs/BETA_TESTING.md"
cp "$ROOT_DIR/BridgeAIChat_PRD.md" "$OUT_DIR/docs/BridgeAIChat_PRD.md"
cp "$ROOT_DIR/CHANGELOG.md" "$OUT_DIR/docs/CHANGELOG.md"

cat > "$OUT_DIR/README.md" <<EOF
# BridgeAIChat $VERSION

## Contents

- \`gateway/\` — gateway binary, built frontend, and \`run-gateway.sh\`
- \`docs/\` — beta setup notes

## Quick Start

1. On the machine hosting the UI and gateway:
   - open \`gateway/\`
   - run \`./run-gateway.sh\`
2. Open \`http://localhost:8080\`
3. Download a matching agent release separately from the future \`bridge-agent\` repository
EOF

(
  cd "$OUT_DIR"
  tar -czf "bridge-app_${VERSION}_darwin_arm64.tar.gz" gateway docs README.md
  shasum -a 256 "bridge-app_${VERSION}_darwin_arm64.tar.gz" > SHA256SUMS.txt
)

echo
echo "Beta bundle ready:"
echo "  $OUT_DIR"
