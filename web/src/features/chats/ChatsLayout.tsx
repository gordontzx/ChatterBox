import { Outlet } from 'react-router-dom'
import { Avatar, Button } from '../../components/ui'
import { ConnectionStatus } from '../../components/ConnectionStatus'
import { useAuth } from '../auth/AuthProvider'

// Real chat list lands in M2 (GET /chats). This is the M0/M1 shell: sidebar
// layout + placeholder rows, now with the real logged-in user.
export function ChatsLayout() {
  const { user, logout } = useAuth()

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
            gap: 8,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
            <Avatar name={user?.name ?? '?'} size={32} />
            <div style={{ minWidth: 0 }}>
              <div
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {user?.name ?? 'ChatterBox'}
              </div>
              <ConnectionStatus />
            </div>
          </div>
          <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
            <Button variant="secondary" disabled>
              + New
            </Button>
            <Button variant="ghost" onClick={() => void logout()}>
              Log out
            </Button>
          </div>
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
