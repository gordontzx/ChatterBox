import type { Chat, Member } from '../../types'

// Direct chats only (backend has no group chats), so "the other member" is
// well-defined. Falls back to the first member if selfId doesn't match
// anyone — shouldn't happen once auth has settled, but keeps this from
// blowing up if it's called during a transient state.
export function otherMember(chat: Chat, selfId: string): Member | undefined {
  return chat.members.find((m) => m.id !== selfId) ?? chat.members[0]
}

export function chatTitle(chat: Chat, selfId: string): string {
  return otherMember(chat, selfId)?.name ?? 'Unknown'
}
