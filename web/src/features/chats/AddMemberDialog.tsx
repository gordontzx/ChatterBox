import { useState, type FormEvent } from 'react'
import { Button, Dialog, Input, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import type { Chat, User } from '../../types'
import { addMember, searchUserByEmail } from './api'
import { useChatsStore } from './chatStore'

interface AddMemberDialogProps {
  open: boolean
  onClose: () => void
  chat: Chat
}

export function AddMemberDialog({ open, onClose, chat }: AddMemberDialogProps) {
  const addMemberLocally = useChatsStore((s) => s.addMemberLocally)

  const [email, setEmail] = useState('')
  const [searching, setSearching] = useState(false)
  const [found, setFound] = useState<User | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [adding, setAdding] = useState(false)

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
      } else if (chat.members.some((m) => m.id === user.id)) {
        toast.error(`${user.name} is already in this chat`)
      } else {
        setFound(user)
      }
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Search failed')
    } finally {
      setSearching(false)
    }
  }

  async function onAdd() {
    if (!found) return
    setAdding(true)
    try {
      await addMember(chat.id, found.id)
      addMemberLocally(chat.id, { id: found.id, name: found.name })
      toast.info(`Added ${found.name}`)
      handleClose()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Could not add member')
    } finally {
      setAdding(false)
    }
  }

  return (
    <Dialog open={open} onClose={handleClose} title="Add someone to this group">
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
            <Button type="button" onClick={() => void onAdd()} disabled={adding}>
              {adding ? 'Adding…' : 'Add'}
            </Button>
          </div>
        )}
      </form>
    </Dialog>
  )
}
