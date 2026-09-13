import { create } from 'zustand'
import { ApiError } from '../../lib/apiClient'
import { fetchReadReceipts } from './api'

// A fresh `{}`/`[]` literal in a zustand selector is a new reference on
// every call, which useSyncExternalStore treats as a changed snapshot and
// loops forever (bit TypingIndicator earlier this project). Stable empty
// constants for callers that do `receiptsFor(chatId) ?? EMPTY`.
export const EMPTY_RECEIPTS: Record<string, number> = {}

interface ChatReadState {
  receipts: Record<string, number> // userId -> last_read_seq
  loaded: boolean
  loading: boolean
}

interface ReadState {
  chats: Record<string, ChatReadState>
  // Seeds a chat's receipts once from GET .../read_receipts. Guarded like
  // messageStore.ensureLoaded — call refreshAll() to force a re-fetch (e.g.
  // on WS reconnect, to catch whatever was missed while disconnected).
  ensureLoaded: (chatId: string) => Promise<void>
  refreshAll: () => void
  // From the WS "read" event — non-regressing, mirrors the backend's
  // GREATEST() so an out-of-order delivery can't move a tick backward.
  markRead: (chatId: string, userId: string, upToSeq: number) => void
  reset: () => void
}

function emptyChat(): ChatReadState {
  return { receipts: {}, loaded: false, loading: false }
}

export const useReadStore = create<ReadState>((set, get) => ({
  chats: {},

  ensureLoaded: async (chatId) => {
    const current = get().chats[chatId]
    if (current?.loaded || current?.loading) return

    set((s) => ({
      chats: { ...s.chats, [chatId]: { ...emptyChat(), ...current, loading: true } },
    }))

    try {
      const receipts = await fetchReadReceipts(chatId)
      const byUser: Record<string, number> = {}
      for (const r of receipts) byUser[r.user_id] = r.last_read_seq
      set((s) => ({
        chats: { ...s.chats, [chatId]: { receipts: byUser, loaded: true, loading: false } },
      }))
    } catch (err) {
      // Best-effort feature — if this fails, ticks just don't show up
      // rather than blocking the conversation from loading.
      set((s) => ({
        chats: { ...s.chats, [chatId]: { ...emptyChat(), ...s.chats[chatId], loading: false } },
      }))
      if (err instanceof ApiError) {
        console.error('failed to load read receipts:', err.message)
      }
    }
  },

  refreshAll: () => {
    for (const chatId of Object.keys(get().chats)) {
      set((s) => ({
        chats: { ...s.chats, [chatId]: { ...s.chats[chatId], loaded: false } },
      }))
      void get().ensureLoaded(chatId)
    }
  },

  markRead: (chatId, userId, upToSeq) => {
    set((s) => {
      const chat = s.chats[chatId] ?? emptyChat()
      const current = chat.receipts[userId] ?? 0
      if (upToSeq <= current) return s
      return {
        chats: {
          ...s.chats,
          [chatId]: { ...chat, receipts: { ...chat.receipts, [userId]: upToSeq } },
        },
      }
    })
  },

  reset: () => set({ chats: {} }),
}))
