import type { ButtonHTMLAttributes, AnchorHTMLAttributes, ReactNode } from 'react'
import './Button.css'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'
type ButtonSize = 'sm' | 'md' | 'lg'

interface BaseButtonProps {
  children: ReactNode
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
}

type NativeButtonProps = BaseButtonProps & ButtonHTMLAttributes<HTMLButtonElement> & {
  href?: never
}

type AnchorButtonProps = BaseButtonProps & AnchorHTMLAttributes<HTMLAnchorElement> & {
  href: string
  disabled?: boolean
}

export function Button(props: NativeButtonProps | AnchorButtonProps) {
  const {
    children,
    className = '',
    variant = 'secondary',
    size = 'md',
    loading = false,
    ...rest
  } = props
  const classes = `ui-button ui-button--${variant} ui-button--${size}${loading ? ' is-loading' : ''}${className ? ` ${className}` : ''}`

  if ('href' in props && props.href) {
    const { disabled, href, ...anchorRest } = rest as AnchorButtonProps
    return (
      <a className={`${classes}${disabled ? ' is-disabled' : ''}`} href={disabled ? undefined : href} aria-disabled={disabled || undefined} {...anchorRest}>
        {loading && <span className="ui-button__spinner" aria-hidden="true" />}
        {children}
      </a>
    )
  }

  const buttonRest = rest as ButtonHTMLAttributes<HTMLButtonElement>
  return (
    <button className={classes} disabled={loading || buttonRest.disabled} {...buttonRest}>
      {loading && <span className="ui-button__spinner" aria-hidden="true" />}
      {children}
    </button>
  )
}
