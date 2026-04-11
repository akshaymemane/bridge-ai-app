# BridgeAIChat App

BridgeAIChat is a control-plane app for chatting with AI CLIs running across your own devices over Tailscale.

This repository is the `bridge-app` seed. It owns:

- gateway
- frontend chat UI
- local dev scripts
- release packaging for gateway + UI bundles

This repository does not represent the final community split by itself. The planned public repo layout is:

- `bridge-app` — gateway + UI + app release packaging
- `bridge-agent` — remote device runtime and agent releases
- `bridge-docs` — public website, docs, downloads, and compatibility matrix

Repo seeds for the other two repositories live under [community](/Users/apple/workspace/bridge-ai-chat/community).

## Quick Start

Run everything locally:

```bash
bash run.sh
```

Build an app release bundle:

```bash
bash scripts/package-app-release.sh v0.1.0-beta.1
```

## Release Scope

App release assets from this repo contain:

- gateway binary
- built frontend UI
- `run-gateway.sh`
- release README
- `SHA256SUMS.txt`

They do not contain agent binaries.

## Community Seeds

- Agent repo seed: [community/bridge-agent](/Users/apple/workspace/bridge-ai-chat/community/bridge-agent)
- Docs repo seed: [community/bridge-docs](/Users/apple/workspace/bridge-ai-chat/community/bridge-docs)

## References

- Product requirements: [BridgeAIChat_PRD.md](/Users/apple/workspace/bridge-ai-chat/BridgeAIChat_PRD.md)
- Beta testing notes: [BETA_TESTING.md](/Users/apple/workspace/bridge-ai-chat/BETA_TESTING.md)
- Distribution plan: [community/COMMUNITY_DISTRIBUTION.md](/Users/apple/workspace/bridge-ai-chat/community/COMMUNITY_DISTRIBUTION.md)
