import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Button, Spinner } from '../../components/ui'
import { useAuth } from '../auth/AuthProvider'
import { AddMemberDialog } from '../chats/AddMemberDialog'
import { useChatsStore } from '../chats/chatStore'
import { GroupMembersDialog } from '../chats/GroupMembersDialog'
import { LeaveChatDialog } from '../chats/LeaveChatDialog'
import { chatTitle } from '../chats/utils'
import { useRealtime } from '../realtime/RealtimeProvider'
import { Composer } from './Composer'
import { MessageList } from './MessageList'
import { selectSortedMessages, useMessagesStore } from './messageStore'
import { TypingIndicator } from './TypingIndicator'

export function ConversationPage() {
  const { chatId } = useParams<{ chatId: string }>()
  const { user } = useAuth()

  const chats = useChatsStore((s) => s.chats)
  const chatsLoaded = useChatsStore((s) => s.loaded)
  const chat = chats.find((c) => c.id === chatId)

  const chatMessages = useMessagesStore((s) => (chatId ? s.chats[chatId] : undefined))
  const ensureLoaded = useMessagesStore((s) => s.ensureLoaded)
  const { status: wsStatus, sendMessage, sendTyping } = useRealtime()
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [membersOpen, setMembersOpen] = useState(false)
  const [leaveOpen, setLeaveOpen] = useState(false)

  useEffect(() => {
    if (chatId) void ensureLoaded(chatId)
  }, [chatId, ensureLoaded])

  if (!chatId || !user) return null

  const messages = selectSortedMessages(chatMessages)
  const memberNames = new Map(chat?.members.map((m) => [m.id, m.name]))

  const title = chat ? chatTitle(chat, user.id) : chatsLoaded ? 'Chat not found' : 'Loading…'

  return (
    <>
      <header
        style={{
          padding: '12px 16px',
          borderBottom: '1px solid var(--color-border)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 8,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', minWidth: 0 }}>
          <Link to="/" className="chat-back-link">
            ← Back
          </Link>
          <div style={{ minWidth: 0 }}>
            <div
              style={{
                fontWeight: 600,
                fontSize: 14,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {title}
            </div>
            {chat?.type === 'group' && (
              <button
                type="button"
                onClick={() => setMembersOpen(true)}
                style={{
                  background: 'none',
                  border: 'none',
                  padding: 0,
                  font: 'inherit',
                  fontSize: 12,
                  color: 'var(--color-text-muted)',
                  cursor: 'pointer',
                  textDecoration: 'underline',
                }}
              >
                {chat.members.length} members
              </button>
            )}
          </div>
        </div>
        {chat?.type === 'group' && (
          <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
            <Button variant="secondary" onClick={() => setAddMemberOpen(true)}>
              + Add
            </Button>
            <Button variant="ghost" onClick={() => setLeaveOpen(true)}>
              Leave
            </Button>
          </div>
        )}
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
          key={`messages-${chatId}`}
          messages={messages}
          selfId={user.id}
          senderName={(id) => memberNames.get(id) ?? '?'}
        />
      )}

      <TypingIndicator chatId={chatId} memberName={(id) => memberNames.get(id) ?? '?'} />

      <Composer
        key={`composer-${chatId}`}
        disabled={wsStatus !== 'open'}
        onSend={(content) => sendMessage(chatId, content)}
        onTyping={() => sendTyping(chatId)}
      />

      {chat?.type === 'group' && (
        <>
          <AddMemberDialog
            open={addMemberOpen}
            onClose={() => setAddMemberOpen(false)}
            chat={chat}
          />
          <GroupMembersDialog
            open={membersOpen}
            onClose={() => setMembersOpen(false)}
            chat={chat}
          />
          <LeaveChatDialog open={leaveOpen} onClose={() => setLeaveOpen(false)} chat={chat} />
        </>
      )}
    </>
  )
}
