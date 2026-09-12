import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Dialog, toast } from '../../components/ui'
import { ApiError } from '../../lib/apiClient'
import type { Chat } from '../../types'
import { useMessagesStore } from '../messages/messageStore'
import { leaveChat } from './api'
import { useChatsStore } from './chatStore'
import { chatTitle } from './utils'

interface LeaveChatDialogProps {
  open: boolean
  onClose: () => void
  chat: Chat
}

// Only ever rendered for a group chat (see ConversationPage) — chatTitle's
// group branch doesn't need a real selfId, so '' is fine here.
export function LeaveChatDialog({ open, onClose, chat }: LeaveChatDialogProps) {
  const navigate = useNavigate()
  const removeChatLocally = useChatsStore((s) => s.removeChatLocally)
  const removeMessages = useMessagesStore((s) => s.removeChat)
  const [leaving, setLeaving] = useState(false)

  async function onConfirm() {
    setLeaving(true)
    try {
      await leaveChat(chat.id)
      removeChatLocally(chat.id)
      removeMessages(chat.id)
      onClose()
      navigate('/')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'Could not leave chat')
    } finally {
      setLeaving(false)
    }
  }

  return (
    <Dialog open={open} onClose={onClose} title="Leave group?">
      <p style={{ fontSize: 14, margin: '0 0 16px' }}>
        You&rsquo;ll stop receiving messages from &ldquo;{chatTitle(chat, '')}&rdquo;. You can only
        rejoin if another member adds you back.
      </p>
      <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
        <Button variant="secondary" onClick={onClose} disabled={leaving}>
          Cancel
        </Button>
        <Button variant="danger" onClick={() => void onConfirm()} disabled={leaving}>
          {leaving ? 'Leaving…' : 'Leave'}
        </Button>
      </div>
    </Dialog>
  )
}
