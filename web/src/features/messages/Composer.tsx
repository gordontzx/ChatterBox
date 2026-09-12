import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Button, Input } from '../../components/ui'

const MAX_CONTENT_LENGTH = 4096

interface ComposerProps {
  disabled: boolean
  onSend: (content: string) => void
}

export function Composer({ disabled, onSend }: ComposerProps) {
  const [value, setValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  // A disabled input can't take focus, so plain `autoFocus` would silently
  // fail while the socket is still connecting. Focus explicitly once it's
  // actually usable — covers both "already connected when this chat opens"
  // and "connects a moment after".
  useEffect(() => {
    if (!disabled) inputRef.current?.focus()
  }, [disabled])

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
        ref={inputRef}
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
