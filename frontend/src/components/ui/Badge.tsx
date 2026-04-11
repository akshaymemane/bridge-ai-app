import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

interface BadgeProps {
  variant?: 'online' | 'offline' | 'connecting' | 'error'
  size?: 'sm' | 'md'
  className?: string
  children?: ReactNode
}

const variantClasses = {
  online: 'bg-green-500/20 text-green-400 border border-green-500/30',
  offline: 'bg-gray-500/20 text-gray-400 border border-gray-500/30',
  connecting: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
  error: 'bg-red-500/20 text-red-400 border border-red-500/30',
}

const sizeClasses = {
  sm: 'text-[10px] px-1.5 py-0.5',
  md: 'text-xs px-2 py-1',
}

export function Badge({ variant = 'offline', size = 'sm', className, children }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full font-medium tracking-wide uppercase',
        variantClasses[variant],
        sizeClasses[size],
        className
      )}
    >
      {children}
    </span>
  )
}

/** Dot-only status indicator */
export function StatusDot({ status }: { status: 'online' | 'offline' | 'connecting' }) {
  const colorClass = {
    online: 'bg-green-500',
    offline: 'bg-gray-500',
    connecting: 'bg-yellow-400 animate-pulse',
  }[status]

  return <span className={cn('inline-block w-2 h-2 rounded-full flex-shrink-0', colorClass)} />
}
