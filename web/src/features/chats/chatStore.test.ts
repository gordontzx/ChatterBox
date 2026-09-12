import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Chat } from '../../types'
import { useChatsStore } from './chatStore'

vi.mock('./api', () => ({
  fetchChats: vi.fn(),
}))

import { fetchChats } from './api'

function chat(id: string): Chat {
  return { id, type: 'direct', created_at: '2020-01-01T00:00:00Z', members: [] }
}

describe('chatsStore', () => {
  beforeEach(() => {
    useChatsStore.getState().reset()
    vi.clearAllMocks()
  })

  it('loadChats populates chats and sets loaded', async () => {
    vi.mocked(fetchChats).mockResolvedValueOnce([chat('1')])

    await useChatsStore.getState().loadChats()

    const state = useChatsStore.getState()
    expect(state.loaded).toBe(true)
    expect(state.chats).toHaveLength(1)
    expect(state.error).toBeNull()
  })

  it('loadChats surfaces an error without marking loaded', async () => {
    vi.mocked(fetchChats).mockRejectedValueOnce(new Error('boom'))

    await useChatsStore.getState().loadChats()

    const state = useChatsStore.getState()
    expect(state.loaded).toBe(false)
    expect(state.error).toBeTruthy()
  })

  it('addChat inserts a new chat at the front', () => {
    useChatsStore.setState({ chats: [chat('old')], loaded: true })

    useChatsStore.getState().addChat(chat('new'))

    expect(useChatsStore.getState().chats.map((c) => c.id)).toEqual(['new', 'old'])
  })

  it('addChat updates an existing chat in place instead of duplicating it', () => {
    useChatsStore.setState({ chats: [chat('1')], loaded: true })

    useChatsStore.getState().addChat({ ...chat('1'), created_at: '2021-01-01T00:00:00Z' })

    const state = useChatsStore.getState()
    expect(state.chats).toHaveLength(1)
    expect(state.chats[0].created_at).toBe('2021-01-01T00:00:00Z')
  })

  it('addMemberLocally appends a member without duplicating on repeat calls', () => {
    useChatsStore.setState({ chats: [chat('g1')], loaded: true })

    useChatsStore.getState().addMemberLocally('g1', { id: 'newbie', name: 'Newbie' })
    useChatsStore.getState().addMemberLocally('g1', { id: 'newbie', name: 'Newbie' })

    const members = useChatsStore.getState().chats[0].members
    expect(members).toEqual([{ id: 'newbie', name: 'Newbie' }])
  })

  it('removeChatLocally drops the chat and leaves others untouched', () => {
    useChatsStore.setState({ chats: [chat('1'), chat('2')], loaded: true })

    useChatsStore.getState().removeChatLocally('1')

    expect(useChatsStore.getState().chats.map((c) => c.id)).toEqual(['2'])
  })

  // Regression test for the cross-account leak: reset() is what
  // AuthProvider.clearSessionState() relies on to keep one account's chats
  // from surviving into the next account's session in the same tab.
  it('reset clears everything', () => {
    useChatsStore.setState({ chats: [chat('1')], loaded: true, error: 'x' })

    useChatsStore.getState().reset()

    const state = useChatsStore.getState()
    expect(state.chats).toEqual([])
    expect(state.loaded).toBe(false)
    expect(state.error).toBeNull()
  })
})
