import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Input, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import type { User } from '../../types'
import { useAuth } from '../auth/AuthProvider'
import { createGroupChat, searchUserByEmail } from './api'
import { useChatsStore } from './chatStore'

// Members are added one at a time by exact email (same constraint as
// DirectChatForm — the backend search isn't a typeahead), building up a
// pending list before the group is actually created.
export function GroupChatForm({ onDone }: { onDone: () => void }) {
  const { user: self } = useAuth()
  const addChat = useChatsStore((s) => s.addChat)
  const navigate = useNavigate()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [searching, setSearching] = useState(false)
  const [notFound, setNotFound] = useState(false)
  const [pending, setPending] = useState<User[]>([])
  const [creating, setCreating] = useState(false)

  async function onAddMember(e: FormEvent) {
    e.preventDefault()
    const trimmed = email.trim()
    if (!trimmed) return

    setSearching(true)
    setNotFound(false)
    try {
      const user = await searchUserByEmail(trimmed)
      if (!user) {
        setNotFound(true)
      } else if (self && user.id === self.id) {
        toast.error("That's you!")
      } else if (pending.some((p) => p.id === user.id)) {
        toast.error(`${user.name} is already added`)
      } else {
        setPending((p) => [...p, user])
        setEmail('')
      }
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Search failed')
    } finally {
      setSearching(false)
    }
  }

  function removePending(id: string) {
    setPending((p) => p.filter((u) => u.id !== id))
  }

  async function onCreate() {
    const trimmedName = name.trim()
    if (!trimmedName || pending.length === 0) return

    setCreating(true)
    try {
      const chat = await createGroupChat(
        trimmedName,
        pending.map((u) => u.id),
      )
      addChat(chat)
      onDone()
      navigate(`/chats/${chat.id}`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Could not create group')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      <Input
        placeholder="Group name"
        autoFocus
        value={name}
        onChange={(e) => setName(e.target.value)}
      />

      <form onSubmit={onAddMember} style={{ display: 'flex', gap: 6 }}>
        <Input
          type="email"
          placeholder="Add member by email"
          value={email}
          onChange={(e) => {
            setEmail(e.target.value)
            setNotFound(false)
          }}
          style={{ flex: 1 }}
        />
        <Button type="submit" variant="secondary" disabled={searching || !email.trim()}>
          {searching ? '…' : 'Add'}
        </Button>
      </form>

      {notFound && (
        <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
          No user with that email.
        </p>
      )}

      {pending.length > 0 && (
        <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'grid', gap: 6 }}>
          {pending.map((u) => (
            <li
              key={u.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 8,
                padding: 8,
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-sm)',
                fontSize: 14,
              }}
            >
              {u.name}
              <Button variant="ghost" onClick={() => removePending(u.id)}>
                Remove
              </Button>
            </li>
          ))}
        </ul>
      )}

      <Button
        onClick={() => void onCreate()}
        disabled={creating || !name.trim() || pending.length === 0}
      >
        {creating ? 'Creating…' : 'Create group'}
      </Button>
    </div>
  )
}
