import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Generate a unique chat ID, sanitized to [a-z0-9_-]+
 */
export function generateChatId(): string {
  const raw = `chat_${Date.now()}_${Math.random().toString(36).slice(2)}`
  return raw.replace(/[^a-z0-9_-]/g, '_')
}

/**
 * Derive WebSocket URL from the gateway base URL.
 * http://host → ws://host/ws
 * https://host → wss://host/ws
 */
export function toWsUrl(gatewayUrl: string): string {
  const url = gatewayUrl.replace(/\/$/, '')
  if (url.startsWith('https://')) {
    return url.replace('https://', 'wss://') + '/ws'
  }
  return url.replace('http://', 'ws://') + '/ws'
}

/**
 * Format a timestamp for display in the chat thread.
 */
export function formatTimestamp(ts: number): string {
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
