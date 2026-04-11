import { cn, formatTimestamp } from '../lib/utils'
import type { ChatMessage } from '../types'
import { AlertCircle } from 'lucide-react'

interface MessageBubbleProps {
  message: ChatMessage
}

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user'
  const isError = message.role === 'error'
  const isAssistant = message.role === 'assistant'

  if (isError) {
    return (
      <div className="flex justify-center px-4 py-1 animate-fade-in">
        <div className="flex items-start gap-2 max-w-lg bg-red-950/50 border border-red-500/30 rounded-xl px-4 py-3 text-red-400 text-sm">
          <AlertCircle size={14} className="mt-0.5 shrink-0" />
          <span className="font-mono-chat text-xs break-all">{message.text}</span>
        </div>
      </div>
    )
  }

  return (
    <div
      className={cn(
        'flex w-full px-4 py-1 animate-slide-up',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      {/* Avatar for assistant */}
      {isAssistant && (
        <div className="w-7 h-7 rounded-lg bg-accent/20 border border-accent/30 flex items-center justify-center text-[10px] font-bold text-accent-text shrink-0 mt-1 mr-2">
          AI
        </div>
      )}

      <div
        className={cn(
          'max-w-[75%] flex flex-col gap-1',
          isUser ? 'items-end' : 'items-start'
        )}
      >
        <div
          className={cn(
            'rounded-2xl px-4 py-2.5 text-sm leading-relaxed break-words',
            isUser
              ? 'bg-accent text-white rounded-tr-sm'
              : cn(
                  'bg-surface-3 border border-surface-5 text-gray-200 rounded-tl-sm',
                  'font-mono-chat text-[13px] whitespace-pre-wrap'
                )
          )}
        >
          {message.text || (message.streaming ? '' : <span className="text-gray-500 italic">Empty response</span>)}
          {message.streaming && (
            <span className="streaming-cursor" aria-label="Streaming response" />
          )}
        </div>

        <span className="text-[10px] text-gray-600 px-1">
          {formatTimestamp(message.timestamp)}
        </span>
      </div>

      {/* Avatar placeholder for user (right side spacing) */}
      {isUser && <div className="w-2 shrink-0" />}
    </div>
  )
}
