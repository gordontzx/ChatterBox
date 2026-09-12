import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { Button, Input, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import { useAuth } from './AuthProvider'

const MIN_PASSWORD_LENGTH = 8
const MAX_PASSWORD_LENGTH = 72

export function RegisterPage() {
  const { status, register } = useAuth()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (status === 'authed') {
    return <Navigate to="/" replace />
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    try {
      await register(name, email, password)
      navigate('/', { replace: true })
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Registration failed')
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
        <h1 style={{ fontSize: 18, margin: 0 }}>Create your account</h1>
        <Input
          placeholder="Name"
          autoComplete="name"
          required
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
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
          placeholder={`Password (${MIN_PASSWORD_LENGTH}-${MAX_PASSWORD_LENGTH} characters)`}
          autoComplete="new-password"
          required
          minLength={MIN_PASSWORD_LENGTH}
          maxLength={MAX_PASSWORD_LENGTH}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Button type="submit" disabled={submitting}>
          {submitting ? 'Creating account…' : 'Register'}
        </Button>
        <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
          Already have an account? <Link to="/login">Log in</Link>
        </p>
      </form>
    </div>
  )
}
