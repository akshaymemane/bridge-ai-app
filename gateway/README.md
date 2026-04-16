# gateway

Go gateway for BridgeAIChat. It serves the web UI, accepts browser and agent websocket connections, stores the selected tailnet in the browser session, and fetches devices from the Tailscale API.

## Requirements

- Go 1.22+
- `TAILSCALE_CLIENT_ID`
- `TAILSCALE_CLIENT_SECRET`
- optional `TAILSCALE_API_BASE` override
- `APP_URL` matching the browser origin

## Build

```bash
cd gateway
go build ./cmd/gateway
```

## Run

```bash
export TAILSCALE_CLIENT_ID=tsid_xxx
export TAILSCALE_CLIENT_SECRET=tssecret_xxx
export TAILSCALE_API_BASE=https://api.tailscale.com/api/v2
export APP_URL=http://localhost:8080

go run ./cmd/gateway
```

The gateway listens on `:8080` by default.

## Endpoints

| Endpoint       | Protocol  | Purpose |
|----------------|-----------|---------|
| `/ws`          | WebSocket | Browser UI connects here |
| `/agent`       | WebSocket | Bridge agent connects here |
| `/api/session` | HTTP POST | Stores the selected tailnet in the session |
| `/api/logout`  | HTTP POST | Clears the current session |
| `/api/devices` | HTTP GET  | Returns Tailscale devices merged with live agents |

## `GET /api/devices` response

```json
{
  "devices": [
    {
      "id": "akshays-macbook-pro-2-93d3b409",
      "hostname": "Akshay's MacBook Pro (2)",
      "name": "Akshay's MacBook Pro (2)",
      "os": "macOS",
      "online": true,
      "status": "connected",
      "tailnet_id": "tail427f1a.ts.net"
    }
  ]
}
```

Statuses:

- `connected`: device is online in Tailscale and a Bridge agent is connected
- `agent_missing`: device is online in Tailscale but no Bridge agent is connected
- `offline`: device is offline in Tailscale

## Logs

Structured JSON logs via `log/slog` to stdout. Common fields include `device_id`, `chat_id`, `tailnet_id`, `remote`, and `err`.
