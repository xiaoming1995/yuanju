import type { ReactNode } from 'react'
import { Button } from './Button'
import './ConfirmDialog.css'

interface ConfirmDialogProps {
  open: boolean
  title: ReactNode
  description?: ReactNode
  confirmText?: string
  cancelText?: string
  danger?: boolean
  pending?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  danger = false,
  pending = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  if (!open) return null

  return (
    <div className="ui-confirm-dialog" role="dialog" aria-modal="true" aria-label={typeof title === 'string' ? title : '确认操作'}>
      <div className="ui-confirm-dialog__backdrop" onClick={onCancel} />
      <div className="ui-confirm-dialog__panel">
        <h2 className="ui-confirm-dialog__title serif">{title}</h2>
        {description && <p className="ui-confirm-dialog__description">{description}</p>}
        <div className="ui-confirm-dialog__actions">
          <Button type="button" variant="ghost" onClick={onCancel} disabled={pending}>{cancelText}</Button>
          <Button type="button" variant={danger ? 'danger' : 'primary'} onClick={onConfirm} loading={pending}>
            {confirmText}
          </Button>
        </div>
      </div>
    </div>
  )
}
