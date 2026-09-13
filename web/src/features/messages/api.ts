import { api, orEmpty } from '../../lib/apiClient'
import type { Message, ReadReceipt } from '../../types'

export function fetchMessages(chatId: string, since: number, limit: number): Promise<Message[]> {
  const params = new URLSearchParams({ since: String(since), limit: String(limit) })
  return api.get<Message[] | null>(`/chats/${chatId}/messages?${params}`).then(orEmpty)
}

export function markAsRead(chatId: string, upToSeq: number): Promise<void> {
  return api.post<void>(`/chats/${chatId}/read`, { up_to_seq: upToSeq })
}

export function fetchReadReceipts(chatId: string): Promise<ReadReceipt[]> {
  return api.get<ReadReceipt[] | null>(`/chats/${chatId}/read_receipts`).then(orEmpty)
}
