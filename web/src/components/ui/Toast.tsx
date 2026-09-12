import { create } from 'zustand'

interface ToastItem {
  id: number
  text: string
  tone: 'info' | 'error'
}

interface ToastState {
  toasts: ToastItem[]
  push: (text: string, tone?: ToastItem['tone']) => void
  dismiss: (id: number) => void
}

let nextId = 1

export const useToastStore = create<ToastState>((set) => ({
  toasts: [],
  push: (text, tone = 'info') => {
    const id = nextId++
    set((s) => ({ toasts: [...s.toasts, { id, text, tone }] }))
    setTimeout(() => {
      set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
    }, 4000)
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}))

// Convenience for non-component code (e.g. the API client on a failed request).
export const toast = {
  info: (text: string) => useToastStore.getState().push(text, 'info'),
  error: (text: string) => useToastStore.getState().push(text, 'error'),
}

export function ToastViewport() {
  const toasts = useToastStore((s) => s.toasts)
  const dismiss = useToastStore((s) => s.dismiss)

  if (toasts.length === 0) return null

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 16,
        right: 16,
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        zIndex: 200,
      }}
    >
      {toasts.map((t) => (
        <div
          key={t.id}
          role="status"
          onClick={() => dismiss(t.id)}
          style={{
            cursor: 'pointer',
            background: t.tone === 'error' ? 'var(--color-danger)' : 'var(--color-surface)',
            color: t.tone === 'error' ? '#fff' : 'var(--color-text)',
            border: t.tone === 'error' ? 'none' : '1px solid var(--color-border)',
            borderRadius: 'var(--radius-sm)',
            padding: '10px 14px',
            fontSize: 13,
            boxShadow: '0 4px 12px rgba(0,0,0,0.12)',
            maxWidth: 320,
          }}
        >
          {t.text}
        </div>
      ))}
    </div>
  )
}
