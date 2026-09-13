import { useTypingStore } from './typingStore'

// A fresh `[]` literal in the selector below would be a new reference on
// every call — zustand (via useSyncExternalStore) compares snapshots by
// reference, so that looks like a change on every render and loops forever.
const NO_ONE_TYPING: string[] = []

interface TypingIndicatorProps {
  chatId: string
  memberName: (userId: string) => string
}

export function TypingIndicator({ chatId, memberName }: TypingIndicatorProps) {
  const typingUserIds = useTypingStore((s) => s.typingByChat[chatId] ?? NO_ONE_TYPING)

  if (typingUserIds.length === 0) return null

  const names = typingUserIds.map(memberName)
  const text =
    names.length === 1
      ? `${names[0]} is typing…`
      : names.length === 2
        ? `${names[0]} and ${names[1]} are typing…`
        : `${names.length} people are typing…`

  return (
    <div
      style={{
        padding: '4px 16px',
        fontSize: 12,
        fontStyle: 'italic',
        color: 'var(--color-text-muted)',
      }}
    >
      {text}
    </div>
  )
}
