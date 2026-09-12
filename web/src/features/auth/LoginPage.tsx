import { Button, Input } from '../../components/ui'

// Wiring lands in M1 (AuthProvider + apiClient). This is the M0 shell so
// routing and layout can be verified end-to-end before auth exists.
export function LoginPage() {
  return (
    <div style={{ display: 'grid', placeItems: 'center', minHeight: '100%' }}>
      <form
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
          width: 320,
          background: 'var(--color-surface)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-md)',
          padding: 24,
        }}
      >
        <h1 style={{ fontSize: 18, margin: 0 }}>Log in to ChatterBox</h1>
        <Input type="email" placeholder="Email" autoComplete="email" disabled />
        <Input type="password" placeholder="Password" autoComplete="current-password" disabled />
        <Button type="submit" disabled>
          Log in
        </Button>
        <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
          Don&rsquo;t have an account? <a href="/register">Register</a>
        </p>
      </form>
    </div>
  )
}
