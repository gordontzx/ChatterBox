import { api, orEmpty } from '../../lib/apiClient'
import type { Chat } from '../../types'

export function fetchChats(): Promise<Chat[]> {
  return api.get<Chat[] | null>('/chats').then(orEmpty)
}
