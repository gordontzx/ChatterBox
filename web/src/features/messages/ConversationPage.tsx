import { useParams } from 'react-router-dom'
import { Button, Input } from '../../components/ui'

// Real message history + send lands in M2/M3. This is the M0 shell so the
// header/list/composer layout is settled before wiring data into it.
export function ConversationPage() {
  const { chatId } = useParams<{ chatId: string }>()

  return (
    <>
      <header
        style={{
          padding: '12px 16px',
          borderBottom: '1px solid var(--color-border)',
          fontWeight: 600,
          fontSize: 14,
        }}
      >
        Chat {chatId}
      </header>
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: 16,
          color: 'var(--color-text-muted)',
          fontSize: 13,
        }}
      >
        Message history goes here.
      </div>
      <footer
        style={{
          borderTop: '1px solid var(--color-border)',
          padding: 12,
          display: 'flex',
          gap: 8,
        }}
      >
        <Input placeholder="Message" disabled style={{ flex: 1 }} />
        <Button disabled>Send</Button>
      </footer>
    </>
  )
}
