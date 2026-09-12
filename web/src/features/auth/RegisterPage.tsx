import { Button, Input } from '../../components/ui'

// Wiring lands in M1. See LoginPage for context.
export function RegisterPage() {
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
        <h1 style={{ fontSize: 18, margin: 0 }}>Create your account</h1>
        <Input placeholder="Name" autoComplete="name" disabled />
        <Input type="email" placeholder="Email" autoComplete="email" disabled />
        <Input
          type="password"
          placeholder="Password (8-72 characters)"
          autoComplete="new-password"
          disabled
        />
        <Button type="submit" disabled>
          Register
        </Button>
        <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
          Already have an account? <a href="/login">Log in</a>
        </p>
      </form>
    </div>
  )
}
