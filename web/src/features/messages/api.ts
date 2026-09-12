import { api, orEmpty } from '../../lib/apiClient'
import type { Message } from '../../types'

export function fetchMessages(chatId: string, since: number, limit: number): Promise<Message[]> {
  const params = new URLSearchParams({ since: String(since), limit: String(limit) })
  return api.get<Message[] | null>(`/chats/${chatId}/messages?${params}`).then(orEmpty)
}
