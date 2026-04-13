# bridge-agent

Go bridge agent for BridgeAIChat. Runs on each device, connects to the gateway, and executes AI CLI tools inside persistent tmux sessions.

## Requirements

- Go 1.22+
- `tmux` installed on the device
- One or more AI CLI tools (e.g. `openclaw`, `claude`, `ollama`)
- Optional: `codex` CLI for OpenAI Codex integration

## Build

```bash
cd agent
go mod tidy
go build ./cmd/agent
```

## Configure

```bash
cp ../config/agent.example.yaml agent.yaml
# Edit agent.yaml: set device.id, device.name, tools, gateway.url
```

## Run

```bash
./agent -config ./agent.yaml
# or
go run ./cmd/agent -config ./agent.yaml
```

The `-config` flag defaults to `./agent.yaml` if omitted.

## Startup checks

On startup the agent:

1. Verifies `tmux` is on PATH — exits with a clear error if missing.
2. Warns (but does not abort) for any configured tool binary that is not on PATH.

## Direct mode

Set `direct: true` on any tool that is a one-shot CLI reading from arguments and writing to stdout.

```yaml
tools:
  claude:
    cmd: claude
    args: ["-p"]
    continue_args: ["--continue", "-p"]
    direct: true
```

In direct mode the agent runs the tool as a subprocess, captures stdout when it exits, and streams it back to the gateway. tmux is not used.

**Why this matters:** the tmux pane-scanning path tracks new output by counting visual rows. A 24-row pane with a short response never grows in row count — the sentinel is never detected and the response times out. Direct mode avoids this completely. Use it for all one-shot CLIs (Claude `-p`, any tool that takes input via args and prints a single response).

Tools without `direct: true` use the tmux path, which is still correct for interactive or long-running CLIs that need a persistent shell session.

## Tool fallback

If a tool's binary is missing from PATH, the agent walks a fallback chain rather than failing immediately:

1. Try `fallback_tool` configured on the tool (if set).
2. Try `default_tool` from the top-level config (if different).
3. Error with `tool_not_found` only if every candidate is unreachable.

The same fallback logic applies when the UI requests a tool that is not configured at all.

```yaml
tools:
  codex:
    cmd: codex
    args: ["exec", "--full-auto"]
    continue_args: []
    fallback_tool: openclaw   # try openclaw if codex binary is missing

  openclaw:
    cmd: openclaw
    args: []
    continue_args: []

default_tool: codex           # also used as last-resort fallback
```

Fallback chains are cycle-safe — if a cycle is detected the agent errors immediately.

## Session model

Each `chat_id` maps to a tmux session named `bridge-{chat_id}`.

- Session is created on first message for that `chat_id`.
- Subsequent messages reuse the existing session metadata.
- For CLIs that support explicit conversation resume, configure `continue_args` so follow-up turns can resume prior context.
- `chat_id` must match `[a-z0-9_-]{1,64}`.

Example for Claude CLI:

```yaml
tools:
  claude:
    cmd: claude
    args: ["-p"]
    continue_args: ["--continue", "-p"]
```

Example for Codex CLI:

```yaml
tools:
  codex:
    cmd: codex
    args: ["exec", "--full-auto"]
    continue_args: []
    working_dir: ..
```

Notes for Codex:

- If the workspace is not a git repo, add `--skip-git-repo-check` to `args`.
- `working_dir` is optional and is resolved relative to `agent.yaml` when not absolute.
- Leave `continue_args` empty for now unless you implement explicit per-chat Codex session-id tracking.
- Using `resume --last` is not safe for multi-chat behavior because it can resume the wrong global Codex session.

## Stream termination

1. Agent sends user text to tmux session.
2. Agent immediately sends `echo __BRIDGE_DONE__` as the next command.
3. Agent polls `tmux capture-pane` every 200 ms.
4. When `__BRIDGE_DONE__` appears in the pane output, the agent strips it, emits `stream_end`, and stops polling.
5. A 5-minute timeout kills the polling loop and emits `session_error` if the sentinel never appears.

## Reconnection

On gateway disconnect the agent reconnects with exponential backoff, capped at 30 seconds. On each successful reconnect it re-sends `device_status online`.

## Failure states

| Condition              | Error code         |
|------------------------|--------------------|
| `tmux` not on PATH     | startup abort       |
| Tool binary not found  | `tool_not_found`   |
| Invalid `chat_id`      | `session_error`    |
| tmux session failure   | `session_error`    |
| Response timeout (5m)  | `session_error`    |
