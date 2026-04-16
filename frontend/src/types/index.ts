// ─── Device ────────────────────────────────────────────────────────────────

export interface Device {
  id: string
  device_id: string
  name: string
  hostname?: string
  os?: string
  online?: boolean
  status: 'connected' | 'offline' | 'agent_missing' | 'connecting'
  tools?: string[]
}

// ─── Messages (UI send) ────────────────────────────────────────────────────

export interface SendMessagePayload {
  type: 'send_message'
  chat_id: string
  device_id: string
  tool: string
  text: string
}

// ─── Messages (UI receive) ─────────────────────────────────────────────────

export interface StreamChunkEvent {
  type: 'stream_chunk'
  chat_id: string
  device_id?: string
  user_id?: string
  text: string
}

export interface StreamEndEvent {
  type: 'stream_end'
  chat_id: string
  device_id?: string
  user_id?: string
}

export interface ErrorEvent {
  type: 'error'
  chat_id: string
  device_id?: string
  user_id?: string
  code: string
  message: string
}

export interface DeviceStatusEvent {
  type: 'device_status'
  id?: string
  device_id: string
  name?: string   // present when gateway broadcasts a new device coming online
  hostname?: string
  os?: string
  online?: boolean
  status: 'connected' | 'offline' | 'agent_missing' | 'connecting'
  tools?: string[]
}

export type GatewayEvent =
  | StreamChunkEvent
  | StreamEndEvent
  | ErrorEvent
  | DeviceStatusEvent

// ─── Chat Messages ─────────────────────────────────────────────────────────

export type MessageRole = 'user' | 'assistant' | 'error'

export interface ChatMessage {
  id: string
  role: MessageRole
  text: string
  /** Only on assistant messages — true while chunks are still arriving */
  streaming?: boolean
  timestamp: number
}

// ─── Chat State ────────────────────────────────────────────────────────────

/** Per-device chat state keyed by device_id */
export interface DeviceChat {
  /** The current active chat_id for this device */
  chat_id: string
  messages: ChatMessage[]
}

// ─── WebSocket connection state ───────────────────────────────────────────

export type WsStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface SessionUser {
  user_id: string
  name?: string
  tailnet_id: string
}

export interface AuthSessionResponse {
  authenticated: boolean
  user?: SessionUser
}

export interface DevicesResponse {
  devices: Array<{
    id: string
    hostname?: string
    name: string
    os?: string
    online: boolean
    status: Device['status']
    tools?: string[]
  }>
}
