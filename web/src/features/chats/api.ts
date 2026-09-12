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

export function createGroupChat(name: string, memberIds: string[]): Promise<Chat> {
  return api.post<Chat>('/chats/group', { name, member_ids: memberIds })
}

// 204 No Content — the caller already has the User it resolved via
// searchUserByEmail() and should append it to local state itself
// (chatStore.addMemberLocally); there's no GET /chats/{id} to refetch from.
export function addMember(chatId: string, userId: string): Promise<void> {
  return api.post<void>(`/chats/${chatId}/members`, { user_id: userId })
}
