import './SegmentedTabs.css'

export interface SegmentedTabItem {
  id: string
  label: string
  href?: string
}

interface SegmentedTabsProps {
  items: SegmentedTabItem[]
  activeId?: string
  ariaLabel?: string
  className?: string
}

export function SegmentedTabs({ items, activeId, ariaLabel = '页面分段导航', className = '' }: SegmentedTabsProps) {
  return (
    <nav className={`ui-segmented-tabs${className ? ` ${className}` : ''}`} aria-label={ariaLabel}>
      {items.map((item) => (
        <a
          key={item.id}
          className={`ui-segmented-tabs__item${activeId === item.id ? ' is-active' : ''}`}
          href={item.href || `#${item.id}`}
        >
          {item.label}
        </a>
      ))}
    </nav>
  )
}
