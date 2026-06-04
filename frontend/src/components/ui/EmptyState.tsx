import type { ReactNode } from 'react'
import './EmptyState.css'

interface EmptyStateProps {
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}

export function EmptyState({ title, description, action, className = '' }: EmptyStateProps) {
  return (
    <div className={`ui-empty-state${className ? ` ${className}` : ''}`}>
      <h3 className="ui-empty-state__title serif">{title}</h3>
      {description && <p className="ui-empty-state__description">{description}</p>}
      {action && <div className="ui-empty-state__action">{action}</div>}
    </div>
  )
}
