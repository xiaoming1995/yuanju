import type { HTMLAttributes, ReactNode } from 'react'
import './PageShell.css'

interface PageShellProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  contained?: boolean
}

export function PageShell({ children, className = '', contained = true, ...props }: PageShellProps) {
  return (
    <main className={`ui-page-shell${contained ? ' ui-page-shell--contained' : ''}${className ? ` ${className}` : ''}`} {...props}>
      {children}
    </main>
  )
}
