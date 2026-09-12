import { create } from 'zustand'
import { ApiError } from '../../lib/apiClient'
import type { Chat, Member } from '../../types'
import { fetchChats } from './api'

interface ChatsState {
  chats: Chat[]
  loading: boolean
  loaded: boolean
  error: string | null
  loadChats: () => Promise<void>
  // Inserts a newly-created chat (or updates it if POST /chats/direct
  // returned an existing one) without a full reload.
  addChat: (chat: Chat) => void
  // POST /chats/{id}/members returns 204 with no body, and there's no
  // GET /chats/{id} to refetch from — so the caller (who already resolved
  // the User via searchUserByEmail) appends them to local state directly.
  addMemberLocally: (chatId: string, member: Member) => void
  // Clears everything. Must run on logout (and on a forced sign-out) — this
  // store is a module-level singleton that outlives any one user's session,
  // so without this a second account signing in in the same tab would see
  // the previous account's chats until something happened to trigger a
  // refetch.
  reset: () => void
}

export const useChatsStore = create<ChatsState>((set, get) => ({
  chats: [],
  loading: false,
  loaded: false,
  error: null,
  loadChats: async () => {
    if (get().loading) return
    set({ loading: true, error: null })
    try {
      const chats = await fetchChats()
      set({ chats, loading: false, loaded: true })
    } catch (err) {
      set({ loading: false, error: err instanceof ApiError ? err.message : 'Failed to load chats' })
    }
  },
  addChat: (chat) => {
    set((s) => {
      const exists = s.chats.some((c) => c.id === chat.id)
      return {
        chats: exists ? s.chats.map((c) => (c.id === chat.id ? chat : c)) : [chat, ...s.chats],
      }
    })
  },
  addMemberLocally: (chatId, member) => {
    set((s) => ({
      chats: s.chats.map((c) =>
        c.id === chatId && !c.members.some((m) => m.id === member.id)
          ? { ...c, members: [...c.members, member] }
          : c,
      ),
    }))
  },
  reset: () => set({ chats: [], loading: false, loaded: false, error: null }),
}))
