import { createContext, useContext, useState, useEffect, useCallback, useRef, type ReactNode } from 'react'
import type { Device, DeviceChat, WsStatus, SendMessagePayload } from '../types'
import { useChatState } from '../hooks/useChatState'
import { useWebSocket } from '../hooks/useWebSocket'
import { generateChatId, toWsUrl } from '../lib/utils'

const GATEWAY_URL = (import.meta.env.VITE_GATEWAY_URL as string | undefined) ??
  `${window.location.protocol}//${window.location.host}`

interface AppContextValue {
  // Devices
  devices: Device[]
  devicesLoading: boolean
  devicesError: string | null

  // Selected device
  selectedDeviceId: string | null
  selectDevice: (id: string) => void

  // Chat for selected device
  activeChat: DeviceChat | null

  // Send a message to the selected device
  sendMessage: (text: string) => void

  // WebSocket status
  wsStatus: WsStatus

  // Active chat_id for the selected device (for disabling input while streaming)
  isStreaming: boolean
}

const AppContext = createContext<AppContextValue | null>(null)

export function AppProvider({ children }: { children: ReactNode }) {
  const { state, setDevices, handleGatewayEvent, sendUserMessage } = useChatState()
  const [wsStatus, setWsStatus] = useState<WsStatus>('connecting')
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null)
  const [devicesLoading, setDevicesLoading] = useState(true)
  const [devicesError, setDevicesError] = useState<string | null>(null)

  // Ref to hold current chat_id per device so sendMessage can close over latest value
  const chatIdRef = useRef<Record<string, string>>({})

  const { send } = useWebSocket({
    url: toWsUrl(GATEWAY_URL),
    onEvent: handleGatewayEvent,
    onStatusChange: setWsStatus,
  })

  // Fetch devices on mount
  useEffect(() => {
    const controller = new AbortController()
    const fetchDevices = async () => {
      try {
        const res = await fetch(`${GATEWAY_URL}/devices`, { signal: controller.signal })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as Device[]
        setDevices(data)
        setDevicesError(null)
      } catch (err) {
        if ((err as Error).name === 'AbortError') return
        setDevicesError('Could not load devices. Is the gateway running?')
      } finally {
        setDevicesLoading(false)
      }
    }
    fetchDevices()
    return () => controller.abort()
  }, [setDevices])

  const selectDevice = useCallback((id: string) => {
    setSelectedDeviceId(id)
    // Ensure a chat_id exists for this device
    if (!chatIdRef.current[id]) {
      chatIdRef.current[id] = generateChatId()
    }
  }, [])

  const sendMessage = useCallback(
    (text: string) => {
      if (!selectedDeviceId) return
      // Get or create a chat_id for this device
      if (!chatIdRef.current[selectedDeviceId]) {
        chatIdRef.current[selectedDeviceId] = generateChatId()
      }
      const chat_id = chatIdRef.current[selectedDeviceId]

      // Record user message in local state
      sendUserMessage(selectedDeviceId, chat_id, text)

      // Send to gateway
      const payload: SendMessagePayload = {
        type: 'send_message',
        chat_id,
        device_id: selectedDeviceId,
        tool: '',
        text,
      }
      send(JSON.stringify(payload))
    },
    [selectedDeviceId, sendUserMessage, send]
  )

  // Sync chatIdRef from state (e.g. if state initialised a new chat_id)
  useEffect(() => {
    for (const [deviceId, chat] of Object.entries(state.chats)) {
      if (!chatIdRef.current[deviceId]) {
        chatIdRef.current[deviceId] = chat.chat_id
      }
    }
  }, [state.chats])

  const activeChat = selectedDeviceId ? (state.chats[selectedDeviceId] ?? null) : null

  const isStreaming = Boolean(
    activeChat?.messages.some((m) => m.streaming)
  )

  const value: AppContextValue = {
    devices: state.devices,
    devicesLoading,
    devicesError,
    selectedDeviceId,
    selectDevice,
    activeChat,
    sendMessage,
    wsStatus,
    isStreaming,
  }

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}
