import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Button, Input } from '../../components/ui'

const MAX_CONTENT_LENGTH = 4096

// Send at most one typing frame this often while the user keeps typing,
// rather than one per keystroke — the backend does a membership check plus
// a fan-out query per frame, so unthrottled input would hammer it. Must
// stay below typingStore's TYPING_TIMEOUT_MS (4s) so a continuously-typing
// user's indicator doesn't flicker off between resends.
const TYPING_THROTTLE_MS = 3000

interface ComposerProps {
  disabled: boolean
  onSend: (content: string) => void
  onTyping?: () => void
}

export function Composer({ disabled, onSend, onTyping }: ComposerProps) {
  const [value, setValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const lastTypingSentAtRef = useRef(0)

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

  function handleChange(next: string) {
    setValue(next)
    if (!onTyping || !next) return
    const now = Date.now()
    if (now - lastTypingSentAtRef.current < TYPING_THROTTLE_MS) return
    lastTypingSentAtRef.current = now
    onTyping()
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
        onChange={(e) => handleChange(e.target.value)}
        style={{ flex: 1 }}
      />
      <Button type="submit" disabled={disabled || !value.trim()}>
        Send
      </Button>
    </form>
  )
}
