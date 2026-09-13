import { useEffect, useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { Avatar, Button, Spinner } from '../../components/ui'
import { ConnectionStatus } from '../../components/ConnectionStatus'
import { useAuth } from '../auth/AuthProvider'
import { PresenceDot } from '../presence/PresenceDot'
import { usePresenceStore } from '../presence/presenceStore'
import { useChatsStore } from './chatStore'
import { NewChatDialog } from './NewChatDialog'
import { chatTitle, otherMember } from './utils'

export function ChatsLayout() {
  const { user, logout } = useAuth()
  const { chats, loading, loaded, error, loadChats } = useChatsStore()
  const presenceByUser = usePresenceStore((s) => s.byUser)
  const ensurePresenceLoaded = usePresenceStore((s) => s.ensureLoaded)
  const [newChatOpen, setNewChatOpen] = useState(false)
  const location = useLocation()

  // Below 640px only one pane shows at a time (see .chat-shell in
  // index.css); which one follows the route rather than separate state, so
  // the browser back button naturally goes from conversation back to list.
  const mobileView = location.pathname === '/' ? 'list' : 'conversation'

  useEffect(() => {
    if (!loaded && !loading) void loadChats()
  }, [loaded, loading, loadChats])

  // Seed presence for everyone across every chat, so the online-first sort
  // below has data to work with (each ensureLoaded call is individually
  // guarded, so this is cheap once cached).
  useEffect(() => {
    for (const chat of chats) {
      for (const member of chat.members) {
        if (member.id !== user?.id) void ensurePresenceLoaded(member.id)
      }
    }
  }, [chats, user?.id, ensurePresenceLoaded])

  function isChatOnline(chat: (typeof chats)[number]): boolean {
    return chat.members.some((m) => m.id !== user?.id && presenceByUser[m.id]?.online)
  }

  // Stable sort: online chats bubble to the top, otherwise the backend's
  // existing created_at-desc order is preserved within each group.
  const sortedChats = [...chats].sort((a, b) => Number(isChatOnline(b)) - Number(isChatOnline(a)))

  return (
    <div className="chat-shell" data-mobile-view={mobileView}>
      <aside className="chat-sidebar">
        <header
          style={{
            padding: 16,
            borderBottom: '1px solid var(--color-border)',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 8,
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
              <Avatar name={user?.name ?? '?'} size={32} />
              <div
                style={{
                  minWidth: 0,
                  fontSize: 13,
                  fontWeight: 600,
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {user?.name ?? 'ChatterBox'}
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
          </div>
          {/* Its own row, full width, so it never fights the buttons for
              horizontal space — see bug: it used to share a row with them
              in a fixed 280px sidebar and got visually covered when squeezed. */}
          <div style={{ marginTop: 6, paddingLeft: 42 }}>
            <ConnectionStatus />
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
          {sortedChats.map((chat) => {
            const title = user ? chatTitle(chat, user.id) : '?'
            const other = user && chat.type === 'direct' ? otherMember(chat, user.id) : undefined
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
                <div style={{ position: 'relative', flexShrink: 0 }}>
                  <Avatar name={title} size={32} />
                  {other && <PresenceDot userId={other.id} />}
                </div>
                <span
                  style={{
                    fontSize: 14,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {title}
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
