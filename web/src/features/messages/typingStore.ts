import { create } from 'zustand'

// The backend only ever sends "someone is typing" — there's no "stopped
// typing" frame — so we infer "stopped" from silence: each mark resets a
// per-(chat,user) timer, and the entry clears itself if no new typing
// event arrives before it fires. Keep this comfortably above the
// Composer's resend throttle (see Composer.tsx) so a continuously-typing
// user's indicator doesn't flicker off between resends.
const TYPING_TIMEOUT_MS = 4000

interface TypingState {
  typingByChat: Record<string, string[]> // chatId -> userIds currently "typing"
  markTyping: (chatId: string, userId: string) => void
  reset: () => void
}

// Timer handles aren't serializable UI state, so they live outside the
// store itself — just bookkeeping for when to clear an entry.
const timers = new Map<string, ReturnType<typeof setTimeout>>()

function timerKey(chatId: string, userId: string): string {
  return `${chatId}:${userId}`
}

export const useTypingStore = create<TypingState>((set) => ({
  typingByChat: {},

  markTyping: (chatId, userId) => {
    const key = timerKey(chatId, userId)
    const existingTimer = timers.get(key)
    if (existingTimer) clearTimeout(existingTimer)

    set((s) => {
      const current = s.typingByChat[chatId] ?? []
      if (current.includes(userId)) return s
      return { typingByChat: { ...s.typingByChat, [chatId]: [...current, userId] } }
    })

    timers.set(
      key,
      setTimeout(() => {
        timers.delete(key)
        set((s) => {
          const current = s.typingByChat[chatId]
          if (!current) return s
          return {
            typingByChat: { ...s.typingByChat, [chatId]: current.filter((id) => id !== userId) },
          }
        })
      }, TYPING_TIMEOUT_MS),
    )
  },

  reset: () => {
    for (const timer of timers.values()) clearTimeout(timer)
    timers.clear()
    set({ typingByChat: {} })
  },
}))
