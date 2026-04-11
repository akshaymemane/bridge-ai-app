import { cn } from '../../lib/utils'
import type { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'ghost' | 'icon'
  size?: 'sm' | 'md' | 'lg'
}

const variantClasses = {
  primary:
    'bg-accent hover:bg-accent-dim text-white font-medium shadow-sm transition-colors',
  ghost:
    'bg-transparent hover:bg-surface-4 text-gray-400 hover:text-gray-200 transition-colors',
  icon:
    'bg-transparent hover:bg-surface-4 text-gray-400 hover:text-gray-200 transition-colors rounded-lg',
}

const sizeClasses = {
  sm: 'px-3 py-1.5 text-sm rounded-lg',
  md: 'px-4 py-2 text-sm rounded-lg',
  lg: 'px-5 py-2.5 text-base rounded-xl',
}

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  disabled,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      disabled={disabled}
      className={cn(
        'inline-flex items-center justify-center gap-2 select-none outline-none',
        'focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface-2',
        'disabled:opacity-40 disabled:cursor-not-allowed',
        variantClasses[variant],
        sizeClasses[size],
        className
      )}
    >
      {children}
    </button>
  )
}
