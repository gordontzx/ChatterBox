import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Button, Input, Spinner } from '../../components/ui'
import { useAuth } from '../auth/AuthProvider'
import { useChatsStore } from '../chats/chatStore'
import { otherMember } from '../chats/utils'
import { MessageList } from './MessageList'
import { selectSortedMessages, useMessagesStore } from './messageStore'

export function ConversationPage() {
  const { chatId } = useParams<{ chatId: string }>()
  const { user } = useAuth()

  const chats = useChatsStore((s) => s.chats)
  const chatsLoaded = useChatsStore((s) => s.loaded)
  const chat = chats.find((c) => c.id === chatId)

  const chatMessages = useMessagesStore((s) => (chatId ? s.chats[chatId] : undefined))
  const ensureLoaded = useMessagesStore((s) => s.ensureLoaded)

  useEffect(() => {
    if (chatId) void ensureLoaded(chatId)
  }, [chatId, ensureLoaded])

  if (!chatId || !user) return null

  const messages = selectSortedMessages(chatMessages)
  const memberNames = new Map(chat?.members.map((m) => [m.id, m.name]))

  const title = chat
    ? otherMember(chat, user.id)?.name
    : chatsLoaded
      ? 'Chat not found'
      : 'Loading…'

  return (
    <>
      <header
        style={{
          padding: '12px 16px',
          borderBottom: '1px solid var(--color-border)',
          fontWeight: 600,
          fontSize: 14,
        }}
      >
        {title}
      </header>

      {chatMessages?.loading && messages.length === 0 ? (
        <div style={{ flex: 1, display: 'grid', placeItems: 'center' }}>
          <Spinner />
        </div>
      ) : chatMessages?.error ? (
        <div style={{ flex: 1, display: 'grid', placeItems: 'center' }}>
          <p style={{ color: 'var(--color-danger)', fontSize: 13 }}>{chatMessages.error}</p>
        </div>
      ) : chatMessages?.loaded && messages.length === 0 ? (
        <div
          style={{
            flex: 1,
            display: 'grid',
            placeItems: 'center',
            color: 'var(--color-text-muted)',
            fontSize: 13,
          }}
        >
          No messages yet. Say hi!
        </div>
      ) : (
        <MessageList
          messages={messages}
          selfId={user.id}
          senderName={(id) => memberNames.get(id) ?? '?'}
        />
      )}

      <footer
        style={{
          borderTop: '1px solid var(--color-border)',
          padding: 12,
          display: 'flex',
          gap: 8,
        }}
      >
        <Input placeholder="Message (sending lands in M3)" disabled style={{ flex: 1 }} />
        <Button disabled>Send</Button>
      </footer>
    </>
  )
}
