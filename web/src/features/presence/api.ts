import { api, ApiError } from '../../lib/apiClient'
import type { PresenceInfo } from '../../types'

// Presence is only visible for someone you already share a chat with — the
// backend 403s anyone else. Every caller of this only ever asks about chat
// members, so this should never actually happen, but null (rather than
// throwing) lets presenceStore degrade to "unknown" instead of erroring.
export async function fetchPresence(userId: string): Promise<PresenceInfo | null> {
  try {
    return await api.get<PresenceInfo>(`/users/${userId}/presence`)
  } catch (err) {
    if (err instanceof ApiError && err.status === 403) return null
    throw err
  }
}
