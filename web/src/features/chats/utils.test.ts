import { describe, expect, it } from 'vitest'
import type { Chat } from '../../types'
import { chatTitle, otherMember } from './utils'

function directChat(): Chat {
  return {
    id: 'chat-1',
    type: 'direct',
    created_at: '2020-01-01T00:00:00Z',
    members: [
      { id: 'self', name: 'Self' },
      { id: 'other', name: 'Other' },
    ],
  }
}

function groupChat(name?: string): Chat {
  return {
    id: 'chat-2',
    type: 'group',
    name,
    created_at: '2020-01-01T00:00:00Z',
    members: [
      { id: 'self', name: 'Self' },
      { id: 'a', name: 'A' },
      { id: 'b', name: 'B' },
    ],
  }
}

describe('otherMember', () => {
  it('returns the member that is not selfId', () => {
    expect(otherMember(directChat(), 'self')?.name).toBe('Other')
  })

  it('falls back to the first member if selfId matches no one', () => {
    expect(otherMember(directChat(), 'nobody')?.name).toBe('Self')
  })
})

describe('chatTitle', () => {
  it('titles a direct chat after the other member', () => {
    expect(chatTitle(directChat(), 'self')).toBe('Other')
  })

  it('titles a group chat by its own name, ignoring selfId', () => {
    expect(chatTitle(groupChat('Team Alpha'), 'self')).toBe('Team Alpha')
  })

  it('falls back to a placeholder for a blank/missing group name', () => {
    expect(chatTitle(groupChat(''), 'self')).toBe('Unnamed group')
    expect(chatTitle(groupChat(undefined), 'self')).toBe('Unnamed group')
    expect(chatTitle(groupChat('   '), 'self')).toBe('Unnamed group')
  })
})
