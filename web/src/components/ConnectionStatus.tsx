import { useRealtime } from '../features/realtime/RealtimeProvider'
import type { WsStatus } from '../lib/ws'

const LABEL: Record<WsStatus, string> = {
  open: 'Connected',
  connecting: 'Connecting…',
  reconnecting: 'Reconnecting…',
  closed: 'Offline',
}

const COLOR: Record<WsStatus, string> = {
  open: 'var(--color-online)',
  connecting: '#eab308',
  reconnecting: '#eab308',
  closed: 'var(--color-offline)',
}

export function ConnectionStatus() {
  const { status } = useRealtime()

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
        style={{ width: 8, height: 8, borderRadius: '50%', background: COLOR[status] }}
      />
      {LABEL[status]}
    </div>
  )
}
