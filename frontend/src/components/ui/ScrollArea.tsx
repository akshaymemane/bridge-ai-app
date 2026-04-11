import { cn } from '../../lib/utils'
import type { HTMLAttributes } from 'react'

interface ScrollAreaProps extends HTMLAttributes<HTMLDivElement> {
  className?: string
}

/**
 * A thin wrapper that just applies overflow-y-auto with the custom scrollbar styles.
 */
export function ScrollArea({ className, children, ...props }: ScrollAreaProps) {
  return (
    <div
      {...props}
      className={cn('overflow-y-auto overflow-x-hidden', className)}
    >
      {children}
    </div>
  )
}
