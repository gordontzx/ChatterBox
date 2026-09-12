// Shown at "/" when no conversation is selected.
export function EmptyChatsPage() {
  return (
    <div
      style={{
        flex: 1,
        display: 'grid',
        placeItems: 'center',
        color: 'var(--color-text-muted)',
        fontSize: 14,
      }}
    >
      Select a chat, or start a new one.
    </div>
  )
}
