import { useState, type FormEvent } from 'react'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Button, Input, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import { useAuth } from './AuthProvider'

export function LoginPage() {
  const { status, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (status === 'authed') {
    const from = (location.state as { from?: { pathname: string } } | null)?.from
    return <Navigate to={from?.pathname ?? '/'} replace />
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    try {
      await login(email, password, 'web')
      navigate('/', { replace: true })
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Login failed')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div style={{ display: 'grid', placeItems: 'center', minHeight: '100%' }}>
      <form
        onSubmit={onSubmit}
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
        <Input
          type="email"
          placeholder="Email"
          autoComplete="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          type="password"
          placeholder="Password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Button type="submit" disabled={submitting}>
          {submitting ? 'Logging in…' : 'Log in'}
        </Button>
        <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
          Don&rsquo;t have an account? <Link to="/register">Register</Link>
        </p>
      </form>
    </div>
  )
}
