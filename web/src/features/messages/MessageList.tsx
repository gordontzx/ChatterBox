import { useEffect, useRef } from 'react'
import type { Message } from '../../types'
import { MessageBubble } from './MessageBubble'

interface MessageListProps {
  messages: Message[]
  selfId: string
  senderName: (senderId: string) => string
}

// Always snaps to the newest message. "Stay pinned only if already near the
// bottom" (so reading old history doesn't get yanked away) is M5 polish.
export function MessageList({ messages, selfId, senderName }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: 'end' })
  }, [messages.length])

  return (
    <div style={{ flex: 1, overflowY: 'auto', padding: 16 }}>
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
  )
}
