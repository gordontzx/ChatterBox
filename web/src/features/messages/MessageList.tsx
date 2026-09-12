import { useEffect, useRef, useState } from 'react'
import type { Message } from '../../types'
import { MessageBubble } from './MessageBubble'

const NEAR_BOTTOM_THRESHOLD_PX = 80

interface MessageListProps {
  messages: Message[]
  selfId: string
  senderName: (senderId: string) => string
}

// Auto-scrolls to the newest message only when the viewer was already near
// the bottom (so reading old history doesn't get yanked away by an
// incoming message); otherwise shows a "new messages" pill. Render this
// with `key={chatId}` at the call site so switching chats resets it.
export function MessageList({ messages, selfId, senderName }: MessageListProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const wasNearBottomRef = useRef(true)
  const prevCountRef = useRef(messages.length)
  const [newBelow, setNewBelow] = useState(false)

  function isNearBottom(): boolean {
    const el = containerRef.current
    if (!el) return true
    return el.scrollHeight - el.scrollTop - el.clientHeight < NEAR_BOTTOM_THRESHOLD_PX
  }

  function handleScroll() {
    const nearBottom = isNearBottom()
    wasNearBottomRef.current = nearBottom
    if (nearBottom) setNewBelow(false)
  }

  useEffect(() => {
    const grew = messages.length > prevCountRef.current
    prevCountRef.current = messages.length
    if (!grew) return

    if (wasNearBottomRef.current) {
      bottomRef.current?.scrollIntoView({ block: 'end' })
    } else {
      setNewBelow(true)
    }
  }, [messages.length])

  function scrollToBottom() {
    bottomRef.current?.scrollIntoView({ block: 'end' })
    setNewBelow(false)
  }

  return (
    <div style={{ position: 'relative', flex: 1, minHeight: 0 }}>
      <div
        ref={containerRef}
        onScroll={handleScroll}
        style={{ height: '100%', overflowY: 'auto', padding: 16 }}
      >
        {messages.map((m) => (
          <MessageBubble
            key={m.id}
            message={m}
            isOwn={m.sender_id === selfId}
            senderName={senderName(m.sender_id)}
          />
        ))}
        <div ref={bottomRef} />
      </div>
      {newBelow && (
        <button
          onClick={scrollToBottom}
          style={{
            position: 'absolute',
            bottom: 12,
            left: '50%',
            transform: 'translateX(-50%)',
            border: 'none',
            borderRadius: 999,
            padding: '6px 14px',
            fontSize: 12,
            background: 'var(--color-primary)',
            color: 'var(--color-primary-text)',
            cursor: 'pointer',
            boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
          }}
        >
          New messages ↓
        </button>
      )}
    </div>
  )
}
