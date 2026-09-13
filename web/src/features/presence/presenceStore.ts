import { create } from 'zustand'
import { ApiError } from '../../lib/apiClient'
import { fetchPresence } from './api'

interface UserPresence {
  online: boolean
  lastSeenAt?: string
  loaded: boolean
  loading: boolean
}

interface PresenceState {
  byUser: Record<string, UserPresence>
  // Fetches a user's current presence once (guarded like messageStore's
  // ensureLoaded); after that, updates arrive live via the WS "presence"
  // event (setPresence). No per-key fallback default in selectors reading
  // byUser[id] — `undefined` for a not-yet-loaded user is referentially
  // stable across renders, unlike a fresh `?? {}`/`?? []` literal would be
  // (that's what caused TypingIndicator's infinite-loop bug earlier).
  ensureLoaded: (userId: string) => Promise<void>
  // Force a re-fetch of every user currently tracked — call on WS reconnect
  // so a presence change missed while disconnected isn't stuck stale
  // (same gap readStore.refreshAll() closes for read receipts).
  refreshAll: () => void
  setPresence: (userId: string, online: boolean, lastSeenAt?: string) => void
  reset: () => void
}

export const usePresenceStore = create<PresenceState>((set, get) => ({
  byUser: {},

  ensureLoaded: async (userId) => {
    const current = get().byUser[userId]
    if (current?.loaded || current?.loading) return

    set((s) => ({
      byUser: { ...s.byUser, [userId]: { online: false, loaded: false, loading: true } },
    }))

    try {
      const info = await fetchPresence(userId)
      set((s) => ({
        byUser: {
          ...s.byUser,
          [userId]: {
            online: info?.online ?? false,
            lastSeenAt: info?.last_seen_at,
            loaded: true,
            loading: false,
          },
        },
      }))
    } catch (err) {
      // Best-effort — if this fails, the indicator just doesn't show up.
      set((s) => ({
        byUser: { ...s.byUser, [userId]: { online: false, loaded: true, loading: false } },
      }))
      if (err instanceof ApiError) {
        console.error('failed to load presence:', err.message)
      }
    }
  },

  refreshAll: () => {
    for (const userId of Object.keys(get().byUser)) {
      set((s) => ({ byUser: { ...s.byUser, [userId]: { ...s.byUser[userId], loaded: false } } }))
      void get().ensureLoaded(userId)
    }
  },

  setPresence: (userId, online, lastSeenAt) => {
    set((s) => ({
      byUser: { ...s.byUser, [userId]: { online, lastSeenAt, loaded: true, loading: false } },
    }))
  },

  reset: () => set({ byUser: {} }),
}))
