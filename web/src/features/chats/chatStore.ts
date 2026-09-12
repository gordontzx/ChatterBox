import { create } from 'zustand'
import { ApiError } from '../../lib/apiClient'
import type { Chat } from '../../types'
import { fetchChats } from './api'

interface ChatsState {
  chats: Chat[]
  loading: boolean
  loaded: boolean
  error: string | null
  loadChats: () => Promise<void>
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
}))
