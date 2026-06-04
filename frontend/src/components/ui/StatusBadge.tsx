import type { ReactNode } from 'react'
import './StatusBadge.css'

type StatusTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral'

interface StatusBadgeProps {
  children: ReactNode
  tone?: StatusTone
  className?: string
}

export function StatusBadge({ children, tone = 'neutral', className = '' }: StatusBadgeProps) {
  return (
    <span className={`ui-status-badge ui-status-badge--${tone}${className ? ` ${className}` : ''}`}>
      {children}
    </span>
  )
}
