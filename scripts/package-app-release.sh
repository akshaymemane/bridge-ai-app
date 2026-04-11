#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-v0.1.0-beta.1}"
OUT_DIR="$ROOT_DIR/dist/bridge-app_${VERSION}"
mkdir -p "$OUT_DIR"

echo "Packaging bridge-app $VERSION"
echo "Output: $OUT_DIR"

(
  cd "$ROOT_DIR/frontend"
  npm run build
)

build_target() {
  local goos="$1"
  local goarch="$2"
  local stage_dir="$OUT_DIR/bridge-app_${VERSION}_${goos}_${goarch}"

  mkdir -p "$stage_dir"

  (
    cd "$ROOT_DIR/gateway"
    GOOS="$goos" GOARCH="$goarch" go build -o "$stage_dir/bridge-gateway" ./cmd/gateway
  )

  cp -R "$ROOT_DIR/frontend/dist" "$stage_dir/ui"

  cat > "$stage_dir/run-gateway.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$DIR/bridge-gateway" -ui-dist "$DIR/ui"
EOF
  chmod +x "$stage_dir/run-gateway.sh"

  cat > "$stage_dir/README.md" <<EOF
# bridge-app $VERSION

This release contains the BridgeAIChat gateway and built frontend UI.

## Run

\`\`\`bash
./run-gateway.sh
\`\`\`

Then open:

\`\`\`text
http://localhost:8080
\`\`\`

Download a matching bridge-agent release separately from the \`bridge-agent\` repository.
EOF
}

build_target darwin arm64
build_target darwin amd64
build_target linux amd64
build_target linux arm64

(
  cd "$OUT_DIR"
  tar -czf "bridge-app_${VERSION}_darwin_arm64.tar.gz" "bridge-app_${VERSION}_darwin_arm64"
  tar -czf "bridge-app_${VERSION}_darwin_amd64.tar.gz" "bridge-app_${VERSION}_darwin_amd64"
  tar -czf "bridge-app_${VERSION}_linux_amd64.tar.gz" "bridge-app_${VERSION}_linux_amd64"
  tar -czf "bridge-app_${VERSION}_linux_arm64.tar.gz" "bridge-app_${VERSION}_linux_arm64"
  shasum -a 256 ./*.tar.gz > SHA256SUMS.txt
)

echo "App release ready:"
echo "  $OUT_DIR"
