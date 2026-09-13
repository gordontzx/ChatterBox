import { useEffect } from 'react'
import { usePresenceStore } from './presenceStore'

interface PresenceDotProps {
  userId: string
}

// Self-contained: fetches its own user's presence (guarded, so mounting
// several of these for a member list is cheap) and only renders while
// online — no dot at all when offline/not-yet-loaded, same "absence is the
// default state" convention as TypingIndicator.
export function PresenceDot({ userId }: PresenceDotProps) {
  const ensureLoaded = usePresenceStore((s) => s.ensureLoaded)
  const presence = usePresenceStore((s) => s.byUser[userId])

  useEffect(() => {
    void ensureLoaded(userId)
  }, [userId, ensureLoaded])

  if (!presence?.online) return null

  return (
    <span
      aria-label="Online"
      style={{
        position: 'absolute',
        bottom: -1,
        right: -1,
        width: 9,
        height: 9,
        borderRadius: '50%',
        background: 'var(--color-online)',
        border: '2px solid var(--color-surface)',
      }}
    />
  )
}
