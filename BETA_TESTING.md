# BridgeAIChat Beta Testing

This guide is for testing BridgeAIChat across a few of your own devices.

## Beta Scope

Current beta is suitable for:

- one operator
- trusted devices on your own Tailscale network
- one gateway host
- multiple agents on other devices
- local AI CLIs such as Codex or Claude where available

This beta is not hardened for:

- multi-user access
- internet exposure
- auth or permissions boundaries
- production reliability

## Recommended Topology

Use one machine as the control host:

- runs the gateway
- serves the web UI
- reachable from your devices over Tailscale

Each additional device runs:

- `bridge-agent`
- local AI CLI(s)
- `tmux`

## Quick Start

### 1. Build the beta bundle

From the repo root:

```bash
bash package-beta.sh
```

This creates a versioned folder under `dist/`.

### 2. Start the gateway and UI on the control host

Open the generated bundle and run:

```bash
cd dist/<beta-version>/gateway
./run-gateway.sh
```

Then open:

```text
http://localhost:8080
```

The gateway serves the built UI directly in beta mode.

### 3. Copy an agent build to another device

Pick the matching build from:

- `agents/darwin-arm64`
- `agents/darwin-amd64`
- `agents/linux-amd64`
- `agents/linux-arm64`

Copy these files to the device:

- `bridge-agent`
- `agent.yaml.example`

Rename the config:

```bash
mv agent.yaml.example agent.yaml
```

### 4. Edit the agent config

Set:

- `device.id`
- `device.name`
- `gateway.url`
- the tool command for the CLI you want on that device

Example:

```yaml
device:
  id: my-laptop
  name: My Laptop

tools:
  codex:
    cmd: codex
    args: ["exec", "--full-auto"]
    continue_args: []
    working_dir: .

default_tool: codex

gateway:
  url: ws://100.x.x.x:8080/agent
```

Use your Tailscale IP or hostname for `gateway.url`.

If the device is not running inside a git repo, add:

```yaml
args: ["exec", "--skip-git-repo-check", "--full-auto"]
```

### 5. Start the agent

On the target device:

```bash
chmod +x bridge-agent
./bridge-agent -config ./agent.yaml
```

Once connected, the device should appear in the BridgeAIChat UI.

## Tool Notes

### Codex

- Works well for this beta as a one-shot executor.
- For non-git folders, include `--skip-git-repo-check`.
- Current beta does not implement per-chat Codex session-id resume yet.

### Claude

- Requires the local Claude CLI to be installed and logged in.
- If Claude hits a quota limit, the UI will not behave like a normal successful reply.

## Operational Notes

- `tmux` is required on agent devices.
- Gateway currently defaults to port `8080`.
- UI and gateway are served from the same binary in beta mode.
- Devices should stay on the same Tailscale network as the gateway.

## Recommended Test Matrix

Try these combinations:

- macOS control host + macOS agent
- macOS control host + Linux VPS agent
- macOS control host + Raspberry Pi agent

For each device, test:

- connect and reconnect
- short prompt
- multi-turn follow-up
- file listing in the working directory
- offline/online transitions

## Known Limitations

- no auth layer
- no installer yet
- no mobile packaging
- no per-chat persisted Codex resume
- limited runtime error surfacing for some CLI failures

## Next Good Improvements After Beta

- installer script for agent devices
- per-chat Codex session-id tracking
- clearer CLI error display in chat
- packaged config wizard
- release archives per platform
