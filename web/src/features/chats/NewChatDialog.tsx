import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Dialog, Input, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import type { User } from '../../types'
import { useAuth } from '../auth/AuthProvider'
import { createDirectChat, searchUserByEmail } from './api'
import { useChatsStore } from './chatStore'

interface NewChatDialogProps {
  open: boolean
  onClose: () => void
}

// Backend user search is exact-match (see BE-2), so this is an
// enter-their-email flow, not a typeahead.
export function NewChatDialog({ open, onClose }: NewChatDialogProps) {
  const { user: self } = useAuth()
  const addChat = useChatsStore((s) => s.addChat)
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [searching, setSearching] = useState(false)
  const [found, setFound] = useState<User | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [starting, setStarting] = useState(false)

  function handleClose() {
    setEmail('')
    setFound(null)
    setNotFound(false)
    onClose()
  }

  async function onSearch(e: FormEvent) {
    e.preventDefault()
    const trimmed = email.trim()
    if (!trimmed) return

    setSearching(true)
    setFound(null)
    setNotFound(false)
    try {
      const user = await searchUserByEmail(trimmed)
      if (!user) {
        setNotFound(true)
      } else if (self && user.id === self.id) {
        toast.error("That's you!")
      } else {
        setFound(user)
      }
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Search failed')
    } finally {
      setSearching(false)
    }
  }

  async function onStart() {
    if (!found) return
    setStarting(true)
    try {
      const chat = await createDirectChat(found.id)
      addChat(chat)
      handleClose()
      navigate(`/chats/${chat.id}`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Could not start chat')
    } finally {
      setStarting(false)
    }
  }

  return (
    <Dialog open={open} onClose={handleClose} title="Start a new chat">
      <form onSubmit={onSearch} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <Input
          type="email"
          placeholder="Their email address"
          autoFocus
          value={email}
          onChange={(e) => {
            setEmail(e.target.value)
            setFound(null)
            setNotFound(false)
          }}
        />
        <Button type="submit" variant="secondary" disabled={searching || !email.trim()}>
          {searching ? 'Searching…' : 'Search'}
        </Button>

        {notFound && (
          <p style={{ fontSize: 13, color: 'var(--color-text-muted)', margin: 0 }}>
            No user with that email.
          </p>
        )}

        {found && (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 8,
              padding: 8,
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-sm)',
            }}
          >
            <span style={{ fontSize: 14 }}>{found.name}</span>
            <Button type="button" onClick={() => void onStart()} disabled={starting}>
              {starting ? 'Starting…' : 'Start chat'}
            </Button>
          </div>
        )}
      </form>
    </Dialog>
  )
}
