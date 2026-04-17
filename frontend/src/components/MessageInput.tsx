import { useState, useRef, type KeyboardEvent } from 'react'
import { Button } from './ui/Button'
import { useApp } from '../context/AppContext'
import { cn, formatToolName } from '../lib/utils'
import { SendHorizonal, Loader2 } from 'lucide-react'

export function MessageInput() {
  const { selectedDeviceId, devices, activeTool, sendMessage, wsStatus, isStreaming } = useApp()
  const [text, setText] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const selectedDevice = devices.find((d) => d.device_id === selectedDeviceId)
  const isUnavailable = selectedDevice?.status !== 'connected'
  const isDisconnected = wsStatus === 'disconnected' || wsStatus === 'error'
  const noDevice = !selectedDeviceId
  const noTool = !activeTool

  const isDisabled = noDevice || noTool || isUnavailable || isDisconnected || isStreaming

  const placeholder = (() => {
    if (noDevice) return 'Select a device to start chatting…'
    if (selectedDevice?.status === 'offline') return 'Device is offline…'
    if (selectedDevice?.status === 'agent_missing') return 'Agent is not installed on this device…'
    if (selectedDevice?.status === 'connecting') return 'Device is still connecting…'
    if (isDisconnected) return 'Reconnecting to gateway…'
    if (isStreaming) return 'AI is responding…'
    if (!activeTool) return 'Choose a tool before sending…'
    return `Message ${selectedDevice?.name ?? 'device'} with ${formatToolName(activeTool)}…`
  })()

  function handleSend() {
    const trimmed = text.trim()
    if (!trimmed || isDisabled) return
    sendMessage(trimmed)
    setText('')
    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  function handleInput() {
    const ta = textareaRef.current
    if (!ta) return
    ta.style.height = 'auto'
    ta.style.height = `${Math.min(ta.scrollHeight, 160)}px`
  }

  return (
    <div className="px-4 pb-4 pt-2 border-t border-surface-5 bg-surface-2">
      <div
        className={cn(
          'flex items-end gap-2 rounded-xl border bg-surface-3 transition-colors',
          isDisabled
            ? 'border-surface-5 opacity-60'
            : 'border-surface-5 focus-within:border-accent/50'
        )}
      >
        <textarea
          ref={textareaRef}
          rows={1}
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          onInput={handleInput}
          disabled={isDisabled}
          placeholder={placeholder}
          className={cn(
            'flex-1 bg-transparent resize-none outline-none text-sm text-gray-200',
            'placeholder:text-gray-600 px-4 py-3',
            'min-h-[44px] max-h-[160px] leading-relaxed',
            'disabled:cursor-not-allowed'
          )}
        />

        <div className="px-2 pb-2 shrink-0">
          <Button
            variant="primary"
            size="sm"
            onClick={handleSend}
            disabled={isDisabled || !text.trim()}
            className="h-8 w-8 p-0"
            aria-label="Send message"
          >
            {isStreaming ? (
              <Loader2 size={14} className="animate-spin" />
            ) : (
              <SendHorizonal size={14} />
            )}
          </Button>
        </div>
      </div>

      <p className="text-[10px] text-gray-600 mt-1.5 px-1">
        {selectedDevice && activeTool && (
          <>
            <span className="mr-2 text-gray-500">Using {formatToolName(activeTool)}</span>
          </>
        )}
        <kbd className="text-[9px] bg-surface-4 border border-surface-5 rounded px-1 py-0.5">Enter</kbd>
        {' '}to send,{' '}
        <kbd className="text-[9px] bg-surface-4 border border-surface-5 rounded px-1 py-0.5">Shift+Enter</kbd>
        {' '}for newline
      </p>
    </div>
  )
}
