# gateway

Go WebSocket gateway for BridgeAIChat.

## Requirements

- Go 1.22+

## Build

```bash
cd gateway
go mod tidy
go build ./cmd/gateway
```

## Run

```bash
./gateway
# or
go run ./cmd/gateway
```

Listens on `:8080` by default.

## Endpoints

| Endpoint    | Protocol  | Purpose                              |
|-------------|-----------|--------------------------------------|
| `/ws`       | WebSocket | Browser UI connects here             |
| `/agent`    | WebSocket | Bridge-agent connects here           |
| `/devices`  | HTTP GET  | Returns JSON list of connected devices |

## GET /devices response

```json
[
  {
    "device_id": "raspberry-pi",
    "name": "Raspberry Pi",
    "status": "online"
  }
]
```

## Failure states

| Condition               | Behaviour                                              |
|-------------------------|--------------------------------------------------------|
| Agent disconnects        | Gateway synthesises `device_status offline` to all UIs |
| Unknown `device_id`      | Error `device_unreachable` sent to requesting UI       |
| Agent send buffer full   | Error `device_unreachable` sent to requesting UI       |
| Invalid JSON from UI     | Error `session_error` sent back to that UI             |

All errors are structured JSON on the WebSocket:

```json
{ "type": "error", "chat_id": "...", "code": "device_unreachable", "message": "..." }
```

Error codes: `device_unreachable`, `tmux_missing`, `tool_not_found`, `session_error`

## Logs

Structured JSON logs via `log/slog` to stdout. Example fields: `device_id`, `chat_id`, `remote`, `err`.
