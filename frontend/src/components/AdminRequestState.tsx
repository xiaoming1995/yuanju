import { Button } from './ui/Button'
import './AdminRequestState.css'

interface Props {
  state: 'loading' | 'error' | 'empty' | 'updating'
  message?: string
  onRetry?: () => void
}

export function AdminRequestState({ state, message, onRetry }: Props) {
  if (state === 'updating') {
    return <div className="admin-request-updating" role="status" aria-busy="true">{message || '正在更新结果…'}</div>
  }
  if (state === 'loading') {
    return <div className="admin-request-loading" role="status" aria-busy="true" aria-label={message || '正在加载'}>
      <div className="admin-request-skeleton" aria-hidden="true" />
      <div className="admin-request-skeleton short" aria-hidden="true" />
      <span>{message || '正在加载…'}</span>
    </div>
  }
  return <div className={`admin-request-state admin-request-${state}`} role={state === 'error' ? 'alert' : 'status'}>
    <p>{message || (state === 'empty' ? '暂无数据' : '加载失败，请重试')}</p>
    {state === 'error' && onRetry && <Button type="button" size="sm" onClick={onRetry}>重试</Button>}
  </div>
}
