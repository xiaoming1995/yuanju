import type { ReactNode } from 'react'
import './FormField.css'

interface FormFieldProps {
  label: ReactNode
  children: ReactNode
  hint?: ReactNode
  error?: ReactNode
  htmlFor?: string
  className?: string
}

export function FormField({ label, children, hint, error, htmlFor, className = '' }: FormFieldProps) {
  return (
    <div className={`ui-form-field${error ? ' has-error' : ''}${className ? ` ${className}` : ''}`}>
      <label className="ui-form-field__label" htmlFor={htmlFor}>{label}</label>
      {children}
      {hint && !error && <p className="ui-form-field__hint">{hint}</p>}
      {error && <p className="ui-form-field__error">{error}</p>}
    </div>
  )
}
