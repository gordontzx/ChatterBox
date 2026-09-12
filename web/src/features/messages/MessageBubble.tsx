import { formatTime } from '../../lib/time'
import type { Message } from '../../types'

interface MessageBubbleProps {
  message: Message
  isOwn: boolean
  senderName: string
}

export function MessageBubble({ message, isOwn, senderName }: MessageBubbleProps) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: isOwn ? 'flex-end' : 'flex-start',
        marginBottom: 10,
      }}
    >
      {!isOwn && (
        <span style={{ fontSize: 11, color: 'var(--color-text-muted)', marginBottom: 2 }}>
          {senderName}
        </span>
      )}
      <div
        style={{
          background: isOwn ? 'var(--color-primary)' : 'var(--color-surface)',
          color: isOwn ? 'var(--color-primary-text)' : 'var(--color-text)',
          border: isOwn ? 'none' : '1px solid var(--color-border)',
          borderRadius: 'var(--radius-md)',
          padding: '8px 12px',
          maxWidth: '70%',
          fontSize: 14,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
        }}
      >
        {message.content}
      </div>
      <span style={{ fontSize: 10, color: 'var(--color-text-muted)', marginTop: 2 }}>
        {formatTime(message.created_at)}
      </span>
    </div>
  )
}
