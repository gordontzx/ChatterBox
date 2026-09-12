import type { Chat, Member } from '../../types'

// Only meaningful for direct chats — "the other member" is well-defined
// there. Falls back to the first member if selfId doesn't match anyone
// (shouldn't happen once auth has settled, but keeps this from blowing up
// if it's called during a transient state).
export function otherMember(chat: Chat, selfId: string): Member | undefined {
  return chat.members.find((m) => m.id !== selfId) ?? chat.members[0]
}

// Display title for any chat: a group chat has its own name; a direct chat
// is titled after whoever you're talking to.
export function chatTitle(chat: Chat, selfId: string): string {
  if (chat.type === 'group') {
    return chat.name?.trim() || 'Unnamed group'
  }
  return otherMember(chat, selfId)?.name ?? 'Unknown'
}
