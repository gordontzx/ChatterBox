import { useEffect } from 'react'
import { formatLastSeen } from '../../lib/time'
import { usePresenceStore } from './presenceStore'

interface PresenceSubtitleProps {
  userId: string
}

// Direct-chat-only (see ConversationPage) — "Online" / "Last seen …" line,
// the same spot TypingIndicator occupies conceptually.
export function PresenceSubtitle({ userId }: PresenceSubtitleProps) {
  const ensureLoaded = usePresenceStore((s) => s.ensureLoaded)
  const presence = usePresenceStore((s) => s.byUser[userId])

  useEffect(() => {
    void ensureLoaded(userId)
  }, [userId, ensureLoaded])

  if (!presence?.loaded) return null

  return (
    <div
      style={{
        fontSize: 12,
        color: presence.online ? 'var(--color-online)' : 'var(--color-text-muted)',
      }}
    >
      {presence.online
        ? 'Online'
        : presence.lastSeenAt
          ? `Last seen ${formatLastSeen(presence.lastSeenAt)}`
          : 'Offline'}
    </div>
  )
}
