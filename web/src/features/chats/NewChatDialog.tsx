import { useState } from 'react'
import { Button, Dialog } from '../../components/ui'
import { DirectChatForm } from './DirectChatForm'
import { GroupChatForm } from './GroupChatForm'

interface NewChatDialogProps {
  open: boolean
  onClose: () => void
}

type Mode = 'direct' | 'group'

export function NewChatDialog({ open, onClose }: NewChatDialogProps) {
  const [mode, setMode] = useState<Mode>('direct')

  function handleClose() {
    setMode('direct')
    onClose()
  }

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      title={mode === 'direct' ? 'Start a new chat' : 'Create a group'}
    >
      <div style={{ display: 'flex', gap: 6, marginBottom: 12 }}>
        <Button
          variant={mode === 'direct' ? 'primary' : 'secondary'}
          onClick={() => setMode('direct')}
          style={{ flex: 1 }}
        >
          Direct
        </Button>
        <Button
          variant={mode === 'group' ? 'primary' : 'secondary'}
          onClick={() => setMode('group')}
          style={{ flex: 1 }}
        >
          Group
        </Button>
      </div>

      {mode === 'direct' ? (
        <DirectChatForm onDone={handleClose} />
      ) : (
        <GroupChatForm onDone={handleClose} />
      )}
    </Dialog>
  )
}
