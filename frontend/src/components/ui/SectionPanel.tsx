import type { HTMLAttributes, ReactNode } from 'react'
import './SectionPanel.css'

interface SectionPanelProps extends Omit<HTMLAttributes<HTMLElement>, 'title'> {
  title?: ReactNode
  description?: ReactNode
  actions?: ReactNode
  children: ReactNode
}

export function SectionPanel({ title, description, actions, children, className = '', ...props }: SectionPanelProps) {
  return (
    <section className={`ui-section-panel${className ? ` ${className}` : ''}`} {...props}>
      {(title || description || actions) && (
        <div className="ui-section-panel__header">
          <div>
            {title && <h2 className="ui-section-panel__title serif">{title}</h2>}
            {description && <p className="ui-section-panel__description">{description}</p>}
          </div>
          {actions && <div className="ui-section-panel__actions">{actions}</div>}
        </div>
      )}
      <div className="ui-section-panel__body">{children}</div>
    </section>
  )
}
