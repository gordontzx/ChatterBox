import { useEffect, useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { Avatar, Button, Spinner } from '../../components/ui'
import { ConnectionStatus } from '../../components/ConnectionStatus'
import { useAuth } from '../auth/AuthProvider'
import { useChatsStore } from './chatStore'
import { NewChatDialog } from './NewChatDialog'
import { otherMember } from './utils'

export function ChatsLayout() {
  const { user, logout } = useAuth()
  const { chats, loading, loaded, error, loadChats } = useChatsStore()
  const [newChatOpen, setNewChatOpen] = useState(false)
  const location = useLocation()

  // Below 640px only one pane shows at a time (see .chat-shell in
  // index.css); which one follows the route rather than separate state, so
  // the browser back button naturally goes from conversation back to list.
  const mobileView = location.pathname === '/' ? 'list' : 'conversation'

  useEffect(() => {
    if (!loaded && !loading) void loadChats()
  }, [loaded, loading, loadChats])

  return (
    <div className="chat-shell" data-mobile-view={mobileView}>
      <aside className="chat-sidebar">
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
            <Button variant="secondary" onClick={() => setNewChatOpen(true)}>
              + New
            </Button>
            <Button variant="ghost" onClick={() => void logout()}>
              Log out
            </Button>
          </div>
        </header>
        <div style={{ flex: 1, overflowY: 'auto', padding: 8 }}>
          {loading && !loaded && (
            <div style={{ display: 'grid', placeItems: 'center', padding: 24 }}>
              <Spinner />
            </div>
          )}
          {error && (
            <p style={{ padding: 8, fontSize: 13, color: 'var(--color-danger)' }}>{error}</p>
          )}
          {loaded && chats.length === 0 && (
            <p style={{ padding: 8, fontSize: 13, color: 'var(--color-text-muted)' }}>
              No chats yet.
            </p>
          )}
          {chats.map((chat) => {
            const member = user ? otherMember(chat, user.id) : undefined
            return (
              <NavLink
                key={chat.id}
                to={`/chats/${chat.id}`}
                style={({ isActive }) => ({
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '8px',
                  borderRadius: 'var(--radius-sm)',
                  textDecoration: 'none',
                  color: 'var(--color-text)',
                  background: isActive ? 'var(--color-bg)' : 'transparent',
                })}
              >
                <Avatar name={member?.name ?? '?'} size={32} />
                <span
                  style={{
                    fontSize: 14,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {member?.name ?? 'Unknown'}
                </span>
              </NavLink>
            )
          })}
        </div>
      </aside>
      <main className="chat-main">
        <Outlet />
      </main>
      <NewChatDialog open={newChatOpen} onClose={() => setNewChatOpen(false)} />
    </div>
  )
}
