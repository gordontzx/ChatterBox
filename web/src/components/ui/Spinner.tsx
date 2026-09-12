interface SpinnerProps {
  size?: number
}

export function Spinner({ size = 20 }: SpinnerProps) {
  return (
    <span
      role="status"
      aria-label="Loading"
      style={{
        display: 'inline-block',
        width: size,
        height: size,
        borderRadius: '50%',
        border: '2px solid var(--color-border)',
        borderTopColor: 'var(--color-primary)',
        animation: 'spin 0.7s linear infinite',
      }}
    />
  )
}
