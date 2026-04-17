# BridgeAIChat App

BridgeAIChat is the control-plane app for chatting with AI CLIs running across your own devices over Tailscale.

This repository is the publishable `bridge-ai-app` repo. It owns:

- the Go gateway
- the frontend chat UI
- local app dev scripts
- app release packaging

The other beta repos now live independently:

- `bridge-ai-agent` — https://github.com/akshaymemane/bridge-ai-agent
- `bridge-ai-docs` — https://github.com/akshaymemane/bridge-ai-docs

## Beta Flow

The current beta uses a simple Tailscale-backed login:

1. Run the gateway with:
   - `TAILSCALE_CLIENT_ID`
   - `TAILSCALE_CLIENT_SECRET`
   - `TAILSCALE_API_BASE`
   - `APP_URL`
2. Open the UI and enter a tailnet like `example.ts.net`
3. The gateway loads devices from Tailscale and merges them with live Bridge agents
4. Pick an explicit tool in the chat header before sending:
   - `Codex` or `Claude` for AI work
   - `Bridge Helper` for safe read-only checks like `status`, `pwd`, `ls`, `read file <path>`, `tail <path>`, and `processes`

Device states:

- `connected`
- `agent_missing`
- `offline`

## Local Development

Run the gateway and frontend:

```bash
./run.sh
```

Or run them separately:

```bash
./run.sh gateway
./run.sh ui
```

## Release Packaging

Build an app release bundle:

```bash
bash scripts/package-app-release.sh v0.1.0-beta.1
```

Each app release contains:

- `bridge-gateway`
- built frontend UI
- `install-app.sh`
- `run-gateway.sh`
- `SHA256SUMS.txt`

## References

- Product requirements: [BridgeAIChat_PRD.md](/Users/apple/workspace/bridge-ai-chat/BridgeAIChat_PRD.md)
- Beta testing notes: [BETA_TESTING.md](/Users/apple/workspace/bridge-ai-chat/BETA_TESTING.md)
