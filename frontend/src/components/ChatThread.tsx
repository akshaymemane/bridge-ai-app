import { useEffect, useRef } from 'react'
import { MessageBubble } from './MessageBubble'
import { ScrollArea } from './ui/ScrollArea'
import { useApp } from '../context/AppContext'
import { MessageSquareDashed, WifiOff } from 'lucide-react'

export function ChatThread() {
  const { activeChat, selectedDeviceId, devices } = useApp()
  const bottomRef = useRef<HTMLDivElement>(null)

  const messages = activeChat?.messages ?? []
  const lastMessageText = messages[messages.length - 1]?.text

  // Auto-scroll to bottom when new messages arrive or streaming text changes
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, lastMessageText])

  const selectedDevice = devices.find((d) => d.device_id === selectedDeviceId)
  const isOffline = selectedDevice?.status === 'offline'

  // No device selected
  if (!selectedDeviceId) {
    return (
      <div className="flex flex-col items-center justify-center flex-1 text-center px-8 select-none">
        <MessageSquareDashed size={48} className="text-gray-600 mb-4" />
        <h2 className="text-lg font-semibold text-gray-400 mb-2">No device selected</h2>
        <p className="text-sm text-gray-600 max-w-xs">
          Select a device from the sidebar to start chatting with its AI.
        </p>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Offline notice banner */}
      {isOffline && (
        <div className="flex items-center justify-center gap-2 px-4 py-2 bg-gray-800/60 border-b border-surface-5 text-gray-400 text-xs">
          <WifiOff size={12} />
          <span>This device is offline. Messages cannot be sent.</span>
        </div>
      )}

      <ScrollArea className="flex-1 py-4">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full min-h-[200px] text-center px-8 select-none mt-16">
            <div className="w-12 h-12 rounded-xl bg-surface-3 border border-surface-5 flex items-center justify-center mb-4">
              <MessageSquareDashed size={22} className="text-gray-500" />
            </div>
            <p className="text-sm text-gray-500">
              {isOffline
                ? 'Device is offline. Waiting for it to reconnect…'
                : 'Send a message to start the conversation.'}
            </p>
          </div>
        )}

        {messages.map((msg) => (
          <MessageBubble key={msg.id} message={msg} />
        ))}

        {/* Scroll anchor */}
        <div ref={bottomRef} />
      </ScrollArea>
    </div>
  )
}
