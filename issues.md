# Code Review Issues

This file captures the functional issues found during a quick review of the current BridgeAIChat implementation.

## 1. Critical: Agent reconnect can incorrectly mark a live device offline

### Problem

When an agent reconnects, the gateway replaces the old connection with the new one, but the old handler still runs its deferred unregister logic. That unregister can remove the newly registered live connection and broadcast a false offline status.

### Why this matters

- Devices can disappear from the UI immediately after reconnecting.
- Messages may fail even though the agent successfully reconnected.
- The online/offline state becomes unreliable under normal network interruptions.

### Relevant code

- `gateway/cmd/gateway/main.go:80`
- `gateway/cmd/gateway/main.go:96`
- `gateway/cmd/gateway/main.go:421`

### Suggested fix direction

Only unregister the specific connection that is being torn down. Do not delete the current device entry blindly by `device_id` alone. One safe pattern is to unregister only if the stored `AgentConn` still matches the connection being closed.

## 2. High: Persistent tmux session does not preserve AI conversation context

### Problem

The current agent implementation keeps a persistent tmux shell session, but each user message runs the AI CLI as a fresh one-shot subprocess. That means the shell persists, but the AI tool conversation does not.

### Why this matters

- This breaks the PRD expectation of persistent AI session memory.
- Follow-up messages do not actually continue the same AI conversation.
- For tools like `claude -p`, every prompt is effectively stateless.

### Relevant code

- `agent/cmd/agent/main.go:171`
- `agent/cmd/agent/main.go:205`
- `agent/cmd/agent/main.go:230`

### Suggested fix direction

Decide whether V1 should truly support persistent conversational CLIs or explicitly scope V1 to stateless one-shot execution. If persistence is required, the tool process must remain alive per chat/tool session and receive subsequent input inside that same live process, not as a new subprocess each turn.

## 3. Medium: Devices that connect after page load do not appear in the UI

### Problem

The frontend fetches devices only once on mount. Later `device_status` events update existing devices but do not insert newly seen ones into state.

### Why this matters

- A device that comes online after the UI loads is invisible until the user refreshes.
- This makes live device discovery look broken.

### Relevant code

- `frontend/src/context/AppContext.tsx:50`
- `frontend/src/hooks/useChatState.ts:46`

### Suggested fix direction

When handling `device_status`, insert the device into `state.devices` if it does not already exist. To support that cleanly, include the device name in the typed frontend event model and reducer path.

## 4. Medium: Streamed multiline output can lose newline boundaries

### Problem

The agent trims trailing newlines from each streamed chunk before sending it. The frontend then concatenates chunk text directly. If output is split across polling intervals, lines can merge together incorrectly.

### Why this matters

- CLI output formatting can become corrupted.
- Markdown, code blocks, logs, and command output may be harder to read or wrong.

### Relevant code

- `agent/cmd/agent/main.go:290`
- `agent/cmd/agent/main.go:301`
- `frontend/src/hooks/useChatState.ts:114`

### Suggested fix direction

Preserve exact line boundaries in streamed output. Avoid trimming meaningful newlines from intermediate chunks, or include a chunking strategy that guarantees safe reconstruction on the frontend.

## 5. Low: Frontend lint script is broken

### Problem

`npm run lint` fails because ESLint v9 is installed but there is no `eslint.config.js` or equivalent flat config file.

### Why this matters

- The advertised lint check cannot run.
- This reduces confidence in future automated verification.

### Relevant code

- `frontend/package.json:6`

### Suggested fix direction

Either add an ESLint v9 flat config or pin ESLint to a version compatible with the existing setup.

## Verification run

The following checks were run during review:

- `go build ./...` in `gateway` ✅
- `go build ./...` in `agent` ✅
- `npm run build` in `frontend` ✅
- `npm run lint` in `frontend` ❌

Lint failure observed:

```text
ESLint couldn't find an eslint.config.(js|mjs|cjs) file.
From ESLint v9.0.0, the default configuration file is now eslint.config.js.
```

## Notes

- No automated tests were present in the repository at review time.
- Findings above are based on source inspection plus build verification.
