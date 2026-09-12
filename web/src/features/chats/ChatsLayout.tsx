import { Outlet } from 'react-router-dom'
import { Avatar, Button } from '../../components/ui'
import { ConnectionStatus } from '../../components/ConnectionStatus'

// Real chat list lands in M2 (GET /chats). This is the M0 shell: sidebar
// layout + placeholder rows so the two-pane structure is settled early.
export function ChatsLayout() {
  return (
    <div style={{ display: 'flex', height: '100%' }}>
      <aside
        style={{
          width: 280,
          flexShrink: 0,
          borderRight: '1px solid var(--color-border)',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <header
          style={{
            padding: 16,
            borderBottom: '1px solid var(--color-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Avatar name="?" size={32} />
            <div>
              <div style={{ fontSize: 13, fontWeight: 600 }}>ChatterBox</div>
              <ConnectionStatus />
            </div>
          </div>
          <Button variant="secondary" disabled>
            + New
          </Button>
        </header>
        <div style={{ flex: 1, overflowY: 'auto', padding: 8, color: 'var(--color-text-muted)' }}>
          <p style={{ fontSize: 13, padding: '8px 8px' }}>No chats yet.</p>
        </div>
      </aside>
      <main style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <Outlet />
      </main>
    </div>
  )
}
