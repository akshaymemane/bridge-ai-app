import { useEffect, useRef, useCallback } from 'react'
import type { GatewayEvent, WsStatus } from '../types'

const RECONNECT_DELAY_MS = 3000
const MAX_RECONNECT_DELAY_MS = 30000

interface UseWebSocketOptions {
  url: string
  enabled?: boolean
  onEvent: (event: GatewayEvent) => void
  onStatusChange: (status: WsStatus) => void
}

interface UseWebSocketReturn {
  send: (data: string) => void
}

export function useWebSocket({ url, enabled = true, onEvent, onStatusChange }: UseWebSocketOptions): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectDelayRef = useRef(RECONNECT_DELAY_MS)
  const unmountedRef = useRef(false)
  const connectRef = useRef<() => void>(() => {})
  // Queue messages sent while WS is connecting/reconnecting
  const pendingRef = useRef<string[]>([])

  const onEventRef = useRef(onEvent)
  const onStatusChangeRef = useRef(onStatusChange)
  onEventRef.current = onEvent
  onStatusChangeRef.current = onStatusChange

  const scheduleReconnect = useCallback(() => {
    if (unmountedRef.current) return
    if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current)
    reconnectTimerRef.current = setTimeout(() => {
      reconnectDelayRef.current = Math.min(reconnectDelayRef.current * 1.5, MAX_RECONNECT_DELAY_MS)
      connectRef.current()
    }, reconnectDelayRef.current)
  }, [])

  const connect = useCallback(() => {
    if (unmountedRef.current) return
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) return

    onStatusChangeRef.current('connecting')

    let ws: WebSocket
    try {
      ws = new WebSocket(url)
    } catch {
      onStatusChangeRef.current('error')
      scheduleReconnect()
      return
    }

    wsRef.current = ws

    ws.onopen = () => {
      if (unmountedRef.current) { ws.close(); return }
      reconnectDelayRef.current = RECONNECT_DELAY_MS
      onStatusChangeRef.current('connected')
      // Flush any messages that were queued while connecting
      const pending = pendingRef.current.splice(0)
      for (const msg of pending) {
        ws.send(msg)
      }
    }

    ws.onmessage = (event) => {
      if (unmountedRef.current) return
      try {
        const parsed = JSON.parse(event.data as string) as GatewayEvent
        onEventRef.current(parsed)
      } catch {
        // Malformed frame — ignore
      }
    }

    ws.onerror = () => {
      // onerror is always followed by onclose; handle there
    }

    ws.onclose = () => {
      if (unmountedRef.current) return
      wsRef.current = null
      onStatusChangeRef.current('disconnected')
      scheduleReconnect()
    }
  }, [scheduleReconnect, url])

  connectRef.current = connect

  useEffect(() => {
    unmountedRef.current = false
    if (!enabled) {
      onStatusChangeRef.current('disconnected')
      return () => {
        unmountedRef.current = true
      }
    }
    connect()

    return () => {
      unmountedRef.current = true
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current)
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [connect, enabled])

  const send = useCallback((data: string) => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    } else {
      // WS not ready yet — queue and send once connected
      pendingRef.current.push(data)
    }
  }, [])

  return { send }
}
