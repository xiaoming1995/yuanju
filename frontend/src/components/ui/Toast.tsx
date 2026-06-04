import { useCallback, useMemo, useState, type ReactNode } from 'react'
import { ToastContext, type ToastTone } from './toastContext'
import './Toast.css'

interface ToastItem {
  id: number
  message: ReactNode
  tone: ToastTone
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([])

  const showToast = useCallback((message: ReactNode, tone: ToastTone = 'info') => {
    const id = Date.now()
    setItems((current) => [...current, { id, message, tone }])
    window.setTimeout(() => {
      setItems((current) => current.filter((item) => item.id !== id))
    }, 3200)
  }, [])

  const value = useMemo(() => ({ showToast }), [showToast])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="ui-toast" aria-live="polite" aria-atomic="true">
        {items.map((item) => (
          <div key={item.id} className={`ui-toast__item ui-toast__item--${item.tone}`}>
            {item.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
