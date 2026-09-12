import { useState, type FormEvent } from 'react'
import { Button, Input } from '../../components/ui'

const MAX_CONTENT_LENGTH = 4096

interface ComposerProps {
  disabled: boolean
  onSend: (content: string) => void
}

export function Composer({ disabled, onSend }: ComposerProps) {
  const [value, setValue] = useState('')

  function submit() {
    const trimmed = value.trim()
    if (!trimmed || trimmed.length > MAX_CONTENT_LENGTH) return
    onSend(trimmed)
    setValue('')
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    submit()
  }

  return (
    <form
      onSubmit={onSubmit}
      style={{ borderTop: '1px solid var(--color-border)', padding: 12, display: 'flex', gap: 8 }}
    >
      <Input
        placeholder={disabled ? 'Connecting…' : 'Message'}
        disabled={disabled}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        style={{ flex: 1 }}
      />
      <Button type="submit" disabled={disabled || !value.trim()}>
        Send
      </Button>
    </form>
  )
}
