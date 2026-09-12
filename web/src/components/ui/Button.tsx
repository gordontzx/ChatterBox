import type { ButtonHTMLAttributes, CSSProperties } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
}

const base: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  gap: 6,
  borderRadius: 'var(--radius-sm)',
  border: '1px solid transparent',
  padding: '8px 14px',
  fontSize: 14,
  fontWeight: 500,
  transition: 'opacity 0.15s ease',
}

const variants: Record<Variant, CSSProperties> = {
  primary: {
    background: 'var(--color-primary)',
    color: 'var(--color-primary-text)',
  },
  secondary: {
    background: 'var(--color-surface)',
    color: 'var(--color-text)',
    borderColor: 'var(--color-border)',
  },
  ghost: {
    background: 'transparent',
    color: 'var(--color-text)',
  },
  danger: {
    background: 'var(--color-danger)',
    color: '#fff',
  },
}

export function Button({ variant = 'primary', style, disabled, ...rest }: ButtonProps) {
  return (
    <button
      {...rest}
      disabled={disabled}
      style={{
        ...base,
        ...variants[variant],
        opacity: disabled ? 0.55 : 1,
        cursor: disabled ? 'not-allowed' : 'pointer',
        ...style,
      }}
    />
  )
}
