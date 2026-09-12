// Real connection state arrives in M3 (RealtimeProvider). Static "offline"
// dot for now so the sidebar layout is final before the socket exists.
export function ConnectionStatus() {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        fontSize: 12,
        color: 'var(--color-text-muted)',
      }}
    >
      <span
        aria-hidden
        style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--color-offline)' }}
      />
      Offline
    </div>
  )
}
