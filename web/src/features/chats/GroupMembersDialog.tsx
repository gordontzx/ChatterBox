import { Avatar, Dialog } from '../../components/ui'
import type { Chat } from '../../types'

interface GroupMembersDialogProps {
  open: boolean
  onClose: () => void
  chat: Chat
}

// Read-only list — reflects whatever's in chat.members at render time, which
// is only ever refreshed by a fresh GET /chats. It won't pick up someone
// else adding a member without a reload (there's no live membership event).
export function GroupMembersDialog({ open, onClose, chat }: GroupMembersDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} title={`Members (${chat.members.length})`}>
      <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'grid', gap: 8 }}>
        {chat.members.map((m) => (
          <li key={m.id} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Avatar name={m.name} size={28} />
            <span style={{ fontSize: 14 }}>{m.name}</span>
          </li>
        ))}
      </ul>
    </Dialog>
  )
}
