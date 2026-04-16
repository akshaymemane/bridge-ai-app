import { useReducer, useCallback } from 'react'
import type { Device, DeviceChat, ChatMessage, GatewayEvent } from '../types'
import { generateChatId } from '../lib/utils'

// ─── State ────────────────────────────────────────────────────────────────

interface ChatState {
  devices: Device[]
  chats: Record<string, DeviceChat>  // keyed by device_id
}

// ─── Actions ──────────────────────────────────────────────────────────────

type Action =
  | { type: 'SET_DEVICES'; devices: Device[] }
  | { type: 'UPDATE_DEVICE_STATUS'; device_id: string; id?: string; name?: string; hostname?: string; os?: string; online?: boolean; status: Device['status']; tools?: string[] }
  | { type: 'ADD_USER_MESSAGE'; device_id: string; chat_id: string; text: string }
  | { type: 'APPEND_CHUNK'; chat_id: string; text: string }
  | { type: 'FINALIZE_STREAM'; chat_id: string }
  | { type: 'ADD_ERROR_MESSAGE'; chat_id: string; device_id: string; message: string; code: string }

// ─── Helpers ──────────────────────────────────────────────────────────────

function makeId(): string {
  return `${Date.now()}_${Math.random().toString(36).slice(2)}`
}

function getOrCreateChat(chats: Record<string, DeviceChat>, device_id: string): DeviceChat {
  return chats[device_id] ?? { chat_id: generateChatId(), messages: [] }
}

// ─── Reducer ──────────────────────────────────────────────────────────────

function reducer(state: ChatState, action: Action): ChatState {
  switch (action.type) {
    case 'SET_DEVICES': {
      // Merge incoming devices with any existing status updates
      const merged = action.devices.map((d) => {
        const existing = state.devices.find((e) => e.device_id === d.device_id)
        return existing
          ? {
              ...d,
              status: d.status,
              hostname: existing.hostname ?? d.hostname,
              os: existing.os ?? d.os,
              online: existing.online ?? d.online,
              tools: existing.tools ?? d.tools,
            }
          : d
      })
      return { ...state, devices: merged }
    }

    case 'UPDATE_DEVICE_STATUS': {
      const exists = state.devices.some((d) => d.device_id === action.device_id)
      if (!exists) {
        // Device connected after page load — insert it so it appears in the sidebar.
        const newDevice: Device = {
          id: action.id ?? action.device_id,
          device_id: action.device_id,
          name: action.name ?? action.device_id,
          hostname: action.hostname,
          os: action.os,
          online: action.online,
          status: action.status,
          tools: action.tools ?? [],
        }
        return { ...state, devices: [...state.devices, newDevice] }
      }
      return {
        ...state,
        devices: state.devices.map((d) =>
          d.device_id === action.device_id
            ? {
                ...d,
                id: action.id ?? d.id,
                name: action.name ?? d.name,
                hostname: action.hostname ?? d.hostname,
                os: action.os ?? d.os,
                online: action.online ?? d.online,
                status: action.status,
                tools: action.tools ?? d.tools,
              }
            : d
        ),
      }
    }

    case 'ADD_USER_MESSAGE': {
      const existing = getOrCreateChat(state.chats, action.device_id)
      const msg: ChatMessage = {
        id: makeId(),
        role: 'user',
        text: action.text,
        timestamp: Date.now(),
      }
      // Use the provided chat_id (caller just generated it)
      const updated: DeviceChat = {
        chat_id: action.chat_id,
        messages: [...existing.messages, msg],
      }
      return { ...state, chats: { ...state.chats, [action.device_id]: updated } }
    }

    case 'APPEND_CHUNK': {
      // Find which device this chat_id belongs to
      const deviceId = Object.keys(state.chats).find(
        (did) => state.chats[did].chat_id === action.chat_id
      )
      if (!deviceId) return state

      const chat = state.chats[deviceId]
      const messages = [...chat.messages]
      // Find last streaming assistant message
      const lastIdx = messages.map((m) => m.streaming).lastIndexOf(true)
      if (lastIdx === -1) {
        // No streaming message yet — start one
        const msg: ChatMessage = {
          id: makeId(),
          role: 'assistant',
          text: action.text,
          streaming: true,
          timestamp: Date.now(),
        }
        messages.push(msg)
      } else {
        messages[lastIdx] = { ...messages[lastIdx], text: messages[lastIdx].text + action.text }
      }

      return {
        ...state,
        chats: {
          ...state.chats,
          [deviceId]: { ...chat, messages },
        },
      }
    }

    case 'FINALIZE_STREAM': {
      const deviceId = Object.keys(state.chats).find(
        (did) => state.chats[did].chat_id === action.chat_id
      )
      if (!deviceId) return state

      const chat = state.chats[deviceId]
      const messages = chat.messages.map((m) =>
        m.streaming ? { ...m, streaming: false } : m
      )
      return {
        ...state,
        chats: { ...state.chats, [deviceId]: { ...chat, messages } },
      }
    }

    case 'ADD_ERROR_MESSAGE': {
      const deviceId = action.device_id ||
        Object.keys(state.chats).find((did) => state.chats[did].chat_id === action.chat_id)
      if (!deviceId) return state

      const existing = getOrCreateChat(state.chats, deviceId)
      // Finalize any in-progress streaming message first
      const messages = existing.messages.map((m) =>
        m.streaming ? { ...m, streaming: false } : m
      )
      const errMsg: ChatMessage = {
        id: makeId(),
        role: 'error',
        text: `[${action.code}] ${action.message}`,
        timestamp: Date.now(),
      }
      return {
        ...state,
        chats: {
          ...state.chats,
          [deviceId]: { ...existing, messages: [...messages, errMsg] },
        },
      }
    }

    default:
      return state
  }
}

// ─── Hook ─────────────────────────────────────────────────────────────────

const initialState: ChatState = {
  devices: [],
  chats: {},
}

export function useChatState() {
  const [state, dispatch] = useReducer(reducer, initialState)

  const setDevices = useCallback((devices: Device[]) => {
    dispatch({ type: 'SET_DEVICES', devices })
  }, [])

  const handleGatewayEvent = useCallback((event: GatewayEvent) => {
    switch (event.type) {
      case 'device_status':
        dispatch({
          type: 'UPDATE_DEVICE_STATUS',
          device_id: event.device_id,
          id: event.id,
          name: event.name,
          hostname: event.hostname,
          os: event.os,
          online: event.online,
          status: event.status,
          tools: event.tools,
        })
        break

      case 'stream_chunk':
        dispatch({ type: 'APPEND_CHUNK', chat_id: event.chat_id, text: event.text })
        break

      case 'stream_end':
        dispatch({ type: 'FINALIZE_STREAM', chat_id: event.chat_id })
        break

      case 'error':
        dispatch({
          type: 'ADD_ERROR_MESSAGE',
          chat_id: event.chat_id,
          device_id: '',
          message: event.message,
          code: event.code,
        })
        break
    }
  }, [])

  const sendUserMessage = useCallback(
    (device_id: string, chat_id: string, text: string) => {
      dispatch({ type: 'ADD_USER_MESSAGE', device_id, chat_id, text })
    },
    []
  )

  return {
    state,
    setDevices,
    handleGatewayEvent,
    sendUserMessage,
    dispatch,
  }
}
