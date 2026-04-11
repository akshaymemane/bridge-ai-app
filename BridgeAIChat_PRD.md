# BridgeAIChat — Product Requirements Document (PRD)

## 1. Overview

BridgeAIChat is a unified chat interface that allows developers to interact with AI tools running across multiple devices (Raspberry Pi, laptop, desktop, VPS) using a WhatsApp-like UI. Each device appears as a separate chat thread and is accessed securely over Tailscale.

BridgeAIChat communicates with devices through a lightweight bridge-agent that executes local AI CLIs (e.g., Claude, Codex, Ollama, OpenClaw) inside persistent tmux sessions.

No APIs are required — execution is CLI-first.

---

## 2. Problem Statement

Developers often run multiple AI-capable environments:

- Raspberry Pi (OpenClaw / Ollama)
- Laptop (Claude CLI)
- Desktop (local models)
- VPS (Codex / build agents)

Managing them requires:

- SSHing into devices
- switching terminals
- losing context
- managing sessions manually

There is no unified interface to interact with all AI environments.

---

## 3. Goals

### Primary Goals

- Unified chat UI for multiple AI devices
- CLI-first execution (no APIs required)
- Persistent AI sessions
- Secure connectivity via Tailscale
- Multi-tool routing per device
- Streaming responses
- Async job support

### Secondary Goals

- Multi-device orchestration
- file transfer between devices
- AI tool routing
- background execution
- device capability discovery

---

## 4. Non Goals (MVP)

- No cloud AI hosting
- No model management
- No agent framework
- No mobile app (web first)
- No multi-user collaboration

---

## 5. Core Concept

Each device appears as a chat thread.

Example:

BridgeAIChat  
- Raspberry Pi  
- MacBook  
- VPS  

Each chat connects to that device's bridge-agent.

Bridge-agent:
- receives message
- routes to CLI
- runs inside tmux
- streams output
- returns response

---

## 6. Architecture

BridgeAIChat UI  
→ Bridge Gateway (WebSocket)  
→ Tailscale Network  
→ Bridge Agent (device)  
→ tmux session  
→ AI CLI (claude / codex / ollama / openclaw)

---

## 7. Device Model

Example configuration:

```yaml
device:
  id: raspberry-pi
  name: Raspberry Pi
  status: online

tools:
  claude:
    cmd: claude

  openclaw:
    cmd: openclaw

  ollama:
    cmd: ollama run llama3

default: openclaw
```

---

## 8. CLI Routing

Explicit selection

```
/claude explain this code
/codex build UI
/openclaw check solar
```

Mention style

```
@claude fix bug
@codex generate PRD
```

Default fallback

```
Check system logs
```

Uses device default tool.

---

## 9. tmux Session Model

Each tool runs in persistent tmux session.

Examples:

- bridge-pi-claude
- bridge-pi-openclaw
- bridge-mac-codex

Session lifecycle:

chat open → create session  
message → send keys  
output → capture pane

---

## 10. Features

### Multi Device Chat

Devices:

- Raspberry Pi
- MacBook
- VPS

Each device = separate chat.

---

### Multi AI CLI per Device

Raspberry Pi:

- Claude
- OpenClaw
- Ollama

MacBook:

- Claude
- Codex

---

### Persistent Sessions

tmux keeps:

- context
- conversation
- state

---

### Streaming Output

CLI output streamed live:

Thinking...  
Generating...  
Writing file...  
Done  

---

### Async Jobs

Long running:

build project

Returns:

Job started

Later:

Job complete

---

### Tool Switching

UI dropdown:

Raspberry Pi [openclaw]

Switch:

- claude
- openclaw
- ollama

---

### File Outputs

CLI creates file:

PRD.md created

Bridge agent:

- detects file
- uploads
- show download

---

### Device Status

- online
- offline

---

### Background Execution

tmux keeps jobs alive after disconnect.

---

### Multiple Sessions

User can create:

- Pi Claude Chat 1
- Pi Claude Chat 2

Separate memory.

---

## 11. Bridge Agent

Responsibilities:

- connect to gateway
- manage tmux sessions
- route CLI calls
- stream output
- handle jobs
- detect tools
- manage sessions

---

## 12. Dependency Management

Bridge agent requires:

- tmux

Startup flow:

check tmux  
if missing → prompt install  
else continue  

Prompt:

tmux required. Install now?

---

## 13. Installation

User runs:

```
curl bridgeai.dev/install | bash
```

Installer:

- installs bridge-agent
- checks tmux
- prompts install
- registers device
- starts agent

---

## 14. Communication Protocol

Message:

```json
{
  "device": "pi",
  "tool": "claude",
  "message": "Explain architecture"
}
```

Response:

```json
{
  "stream": "Thinking..."
}
```

---

## 15. Tech Stack

Frontend:

- React
- Vite
- Tailwind
- shadcn
- WebSocket

Gateway:

- Go (preferred)
or
- Node

Bridge Agent:

- Go (preferred)
- CLI binary

Device runtime:

- tmux
- bash

Networking:

- Tailscale

---

## 16. MVP Scope

MVP includes:

- device registration
- chat UI
- tmux session per tool
- CLI execution
- streaming output
- tool routing
- default tool
- tmux detection

---

## 17. Future Features

- AI router
- multi device delegation
- file transfer
- voice mode
- mobile apps
- job scheduling
- session sharing

---

## 18. Version Plan

V1

- single device
- single tool
- tmux integration

V2

- multi tool
- routing

V3

- multi device
- async jobs

V4

- orchestration

---

## 19. Summary

BridgeAIChat provides:

- unified AI chat
- multi device control
- CLI-first architecture
- tmux session management
- secure tailscale connectivity

This enables developers to control all AI environments from one interface.
