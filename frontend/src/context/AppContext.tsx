import { createContext, useContext, useState, useEffect, useCallback, useRef, type ReactNode } from 'react'
import type { Device, DeviceChat, WsStatus, SendMessagePayload, AuthSessionResponse, SessionUser, DevicesResponse } from '../types'
import { useChatState } from '../hooks/useChatState'
import { useWebSocket } from '../hooks/useWebSocket'
import { generateChatId, toWsUrl } from '../lib/utils'

const GATEWAY_URL = (import.meta.env.VITE_GATEWAY_URL as string | undefined) ??
  `${window.location.protocol}//${window.location.host}`

interface AppContextValue {
  devices: Device[]
  devicesLoading: boolean
  devicesError: string | null
  authLoading: boolean
  authError: string | null
  isAuthenticated: boolean
  currentUser: SessionUser | null
  login: (tailnet: string) => Promise<void>
  logout: () => Promise<void>
  selectedDeviceId: string | null
  selectDevice: (id: string) => void
  activeChat: DeviceChat | null
  sendMessage: (text: string) => void
  wsStatus: WsStatus
  isStreaming: boolean
}

const AppContext = createContext<AppContextValue | null>(null)

export function AppProvider({ children }: { children: ReactNode }) {
  const { state, setDevices, handleGatewayEvent, sendUserMessage } = useChatState()
  const [wsStatus, setWsStatus] = useState<WsStatus>('connecting')
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null)
  const [devicesLoading, setDevicesLoading] = useState(false)
  const [devicesError, setDevicesError] = useState<string | null>(null)
  const [authLoading, setAuthLoading] = useState(true)
  const [authError, setAuthError] = useState<string | null>(null)
  const [sessionInfo, setSessionInfo] = useState<AuthSessionResponse | null>(null)
  const chatIdRef = useRef<Record<string, string>>({})

  const isAuthenticated = Boolean(sessionInfo?.authenticated)
  const currentUser = sessionInfo?.user ?? null

  const { send } = useWebSocket({
    url: toWsUrl(GATEWAY_URL),
    enabled: isAuthenticated,
    onEvent: handleGatewayEvent,
    onStatusChange: setWsStatus,
  })

  const refreshSession = useCallback(async () => {
    setAuthLoading(true)
    try {
      const res = await fetch(`${GATEWAY_URL}/api/session`, { credentials: 'include' })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = (await res.json()) as AuthSessionResponse
      setSessionInfo(data)
      setAuthError(null)
    } catch {
      setAuthError('Could not load session state.')
      setSessionInfo(null)
    } finally {
      setAuthLoading(false)
    }
  }, [])

  useEffect(() => {
    void refreshSession()
  }, [refreshSession])

  useEffect(() => {
    if (!isAuthenticated) {
      setDevices([])
      setDevicesLoading(false)
      setDevicesError(null)
      setSelectedDeviceId(null)
      return
    }

    const controller = new AbortController()
    const fetchDevices = async () => {
      setDevicesLoading(true)
      try {
        const res = await fetch(`${GATEWAY_URL}/api/devices`, {
          signal: controller.signal,
          credentials: 'include',
        })
        if (res.status === 401) {
          setSessionInfo((current) => current ? { ...current, authenticated: false, user: undefined } : current)
          setDevices([])
          return
        }
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as DevicesResponse
        const devices: Device[] = data.devices.map((device) => ({
          id: device.id,
          device_id: device.id,
          hostname: device.hostname,
          name: device.name,
          os: device.os,
          online: device.online,
          status: device.status,
          tools: device.tools ?? [],
        }))
        setDevices(devices)
        setDevicesError(null)
      } catch (err) {
        if ((err as Error).name === 'AbortError') return
        setDevicesError('Could not load tailnet devices.')
      } finally {
        setDevicesLoading(false)
      }
    }

    void fetchDevices()
    return () => controller.abort()
  }, [isAuthenticated, setDevices])

  useEffect(() => {
    if (!selectedDeviceId && state.devices.length > 0) {
      setSelectedDeviceId(state.devices[0].device_id)
      return
    }
    if (selectedDeviceId && !state.devices.some((device) => device.device_id === selectedDeviceId)) {
      setSelectedDeviceId(state.devices[0]?.device_id ?? null)
    }
  }, [selectedDeviceId, state.devices])

  const login = useCallback(async (tailnet: string) => {
    const res = await fetch(`${GATEWAY_URL}/api/session`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ tailnet }),
    })
    if (!res.ok) {
      throw new Error('Could not store tailnet session.')
    }
    const data = (await res.json()) as AuthSessionResponse
    setSessionInfo(data)
    setAuthError(null)
  }, [])

  const logout = useCallback(async () => {
    await fetch(`${GATEWAY_URL}/api/logout`, {
      method: 'POST',
      credentials: 'include',
    })
    setSessionInfo((current) => current ? {
      ...current,
      authenticated: false,
      user: undefined,
    } : null)
    setDevices([])
    setSelectedDeviceId(null)
  }, [setDevices])

  const selectDevice = useCallback((id: string) => {
    setSelectedDeviceId(id)
    if (!chatIdRef.current[id]) {
      chatIdRef.current[id] = generateChatId()
    }
  }, [])

  const sendMessage = useCallback(
    (text: string) => {
      if (!selectedDeviceId || !isAuthenticated) return
      if (!chatIdRef.current[selectedDeviceId]) {
        chatIdRef.current[selectedDeviceId] = generateChatId()
      }
      const chat_id = chatIdRef.current[selectedDeviceId]

      sendUserMessage(selectedDeviceId, chat_id, text)

      const payload: SendMessagePayload = {
        type: 'send_message',
        chat_id,
        device_id: selectedDeviceId,
        tool: '',
        text,
      }
      send(JSON.stringify(payload))
    },
    [isAuthenticated, selectedDeviceId, sendUserMessage, send]
  )

  useEffect(() => {
    for (const [deviceId, chat] of Object.entries(state.chats)) {
      if (!chatIdRef.current[deviceId]) {
        chatIdRef.current[deviceId] = chat.chat_id
      }
    }
  }, [state.chats])

  const activeChat = selectedDeviceId ? (state.chats[selectedDeviceId] ?? null) : null
  const isStreaming = Boolean(activeChat?.messages.some((message) => message.streaming))

  const value: AppContextValue = {
    devices: state.devices,
    devicesLoading,
    devicesError,
    authLoading,
    authError,
    isAuthenticated,
    currentUser,
    login,
    logout,
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
