import { createContext, type ReactNode } from 'react'

export type ToastTone = 'success' | 'error' | 'info'

export interface ToastContextValue {
  showToast: (message: ReactNode, tone?: ToastTone) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)
