import { api, ApiError, orEmpty } from '../../lib/apiClient'
import type { Chat, User } from '../../types'

export function fetchChats(): Promise<Chat[]> {
  return api.get<Chat[] | null>('/chats').then(orEmpty)
}

// GET /users/search is exact-match, not a prefix search: null means "no
// user with that email" (404), not "keep typing".
export async function searchUserByEmail(email: string): Promise<User | null> {
  try {
    return await api.get<User>(`/users/search?email=${encodeURIComponent(email)}`)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null
    throw err
  }
}

export function createDirectChat(otherUserId: string): Promise<Chat> {
  return api.post<Chat>('/chats/direct', { other_user_id: otherUserId })
}
