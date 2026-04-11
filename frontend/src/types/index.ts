// ─── Device ────────────────────────────────────────────────────────────────

export interface Device {
  device_id: string
  name: string
  status: 'online' | 'offline'
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
  text: string
}

export interface StreamEndEvent {
  type: 'stream_end'
  chat_id: string
}

export interface ErrorEvent {
  type: 'error'
  chat_id: string
  code: string
  message: string
}

export interface DeviceStatusEvent {
  type: 'device_status'
  device_id: string
  name?: string   // present when gateway broadcasts a new device coming online
  status: 'online' | 'offline'
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
