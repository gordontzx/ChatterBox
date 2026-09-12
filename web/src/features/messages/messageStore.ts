import { create } from 'zustand'
import { ApiError } from '../../lib/apiClient'
import type { Message } from '../../types'
import { fetchMessages } from './api'

const PAGE_LIMIT = 100

export interface ChatMessages {
  byId: Record<string, Message>
  lastSeq: number
  loading: boolean
  loaded: boolean
  error: string | null
}

interface MessagesState {
  chats: Record<string, ChatMessages>
  // Loads a chat's full history once, by paging forward from the oldest
  // unreceived message (?since=<lastSeq>) until a short read. See BE-4
  // decision: no backward pagination — first open pages through everything.
  ensureLoaded: (chatId: string) => Promise<void>
  // Merges freshly-arrived messages (e.g. from the WS in M3) into a chat
  // that's already loaded, keyed by id so redelivery can't duplicate them.
  upsertMessage: (chatId: string, message: Message) => void
}

function emptyChat(): ChatMessages {
  return { byId: {}, lastSeq: 0, loading: false, loaded: false, error: null }
}

export const useMessagesStore = create<MessagesState>((set, get) => ({
  chats: {},

  ensureLoaded: async (chatId) => {
    const current = get().chats[chatId]
    if (current?.loaded || current?.loading) return

    set((s) => ({
      chats: { ...s.chats, [chatId]: { ...emptyChat(), ...current, loading: true, error: null } },
    }))

    try {
      const byId: Record<string, Message> = {}
      let since = 0
      let lastSeq = 0
      for (;;) {
        const page = await fetchMessages(chatId, since, PAGE_LIMIT)
        for (const m of page) {
          byId[m.id] = m
          if (m.seq > lastSeq) lastSeq = m.seq
        }
        if (page.length < PAGE_LIMIT) break
        since = lastSeq
      }
      set((s) => ({
        chats: {
          ...s.chats,
          [chatId]: { byId, lastSeq, loading: false, loaded: true, error: null },
        },
      }))
    } catch (err) {
      set((s) => ({
        chats: {
          ...s.chats,
          [chatId]: {
            ...emptyChat(),
            ...s.chats[chatId],
            loading: false,
            error: err instanceof ApiError ? err.message : 'Failed to load messages',
          },
        },
      }))
    }
  },

  upsertMessage: (chatId, message) => {
    set((s) => {
      const chat = s.chats[chatId] ?? emptyChat()
      return {
        chats: {
          ...s.chats,
          [chatId]: {
            ...chat,
            byId: { ...chat.byId, [message.id]: message },
            lastSeq: Math.max(chat.lastSeq, message.seq),
          },
        },
      }
    })
  },
}))

export function selectSortedMessages(state: ChatMessages | undefined): Message[] {
  if (!state) return []
  return Object.values(state.byId).sort((a, b) => a.seq - b.seq)
}
