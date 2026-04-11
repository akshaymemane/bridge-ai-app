# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Start Here

Before any work, read these files in order:
1. `BridgeAIChat_PRD.md` — full product spec
2. `CLAUDE.md` — this file
3. `CLAUDE_EXECUTION_PLAYBOOK.md` — milestone plan, prompt patterns, anti-patterns

## Project Status

Pre-implementation. No code exists yet. Lock architecture decisions before writing code — see the "Important Decisions To Lock Early" section of `CLAUDE_EXECUTION_PLAYBOOK.md`.

## What This Is

**BridgeAIChat** is a unified WhatsApp-style chat interface for interacting with AI CLIs (Claude, Codex, Ollama, OpenClaw) running on multiple remote devices (Raspberry Pi, laptop, VPS) over Tailscale. Execution is CLI-first — no direct AI APIs. Each device runs a **bridge-agent** that manages `tmux` sessions and streams CLI output back to the frontend via WebSocket.

## Planned Tech Stack

| Layer | Tech |
|---|---|
| Frontend | React + Vite + TypeScript + Tailwind + shadcn/ui |
| Gateway | Go (preferred) |
| Bridge Agent | Go (CLI binary, deployed to each device) |
| Transport | WebSocket |
| Networking | Tailscale |
| Session runtime | tmux + bash |

## Architecture

```
BridgeAIChat UI (React/Vite)
  → Bridge Gateway (Go WebSocket server)
    → Tailscale network
      → Bridge Agent (Go binary on each device)
        → tmux session (per tool, per chat)
          → AI CLI (claude / codex / ollama / openclaw)
```

**Key design rules:**
- One tmux session per tool per chat (e.g. `bridge-pi-claude`, `bridge-mac-codex`)
- CLI output is captured from `tmux capture-pane` and streamed back
- Each device is registered with a YAML config listing available tools and a default
- Tool routing: explicit (`/claude`, `@codex`), mention, or device default fallback

## Message Protocol

```json
// Frontend → Gateway → Agent
{ "device": "pi", "tool": "claude", "message": "Explain architecture" }

// Agent → Gateway → Frontend (streamed)
{ "stream": "Thinking..." }
```

## Version Plan (scope discipline)

Stay inside the stated version — do not quietly add later-version features:

| Version | Scope |
|---|---|
| V1 | Single device, single tool, tmux integration, streaming output |
| V2 | Multi-tool routing (`/tool`, `@tool`, default fallback) |
| V3 | Multi-device support, async jobs |
| V4 | Orchestration |

## Subagents

Three specialized agents live in `.claude/agents/`. Use them instead of working generically:

| Agent | When to use |
|---|---|
| `bridge-product-architect` | Before broad implementation, when requirements feel fuzzy, scope/risk review, API contract design |
| `bridge-systems-engineer` | Go gateway, bridge-agent, tmux integration, streaming protocol, install/runtime behavior |
| `bridge-ui-engineer` | React/Vite/Tailwind/shadcn UI, chat UX, streaming rendering, frontend state |

See `CLAUDE_EXECUTION_PLAYBOOK.md` for ready-to-use prompt patterns for each agent.
