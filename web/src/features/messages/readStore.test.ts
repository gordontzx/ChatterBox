import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ReadReceipt } from '../../types'
import { useReadStore } from './readStore'

vi.mock('./api', () => ({
  fetchReadReceipts: vi.fn(),
}))

import { fetchReadReceipts } from './api'

function receipt(userId: string, seq: number): ReadReceipt {
  return { user_id: userId, last_read_seq: seq }
}

describe('readStore', () => {
  beforeEach(() => {
    useReadStore.getState().reset()
    vi.clearAllMocks()
  })

  it('ensureLoaded seeds receipts as a userId -> seq map', async () => {
    vi.mocked(fetchReadReceipts).mockResolvedValueOnce([receipt('alice', 5), receipt('bob', 3)])

    await useReadStore.getState().ensureLoaded('chat-1')

    const state = useReadStore.getState().chats['chat-1']
    expect(state.loaded).toBe(true)
    expect(state.receipts).toEqual({ alice: 5, bob: 3 })
  })

  it('does not refetch a chat that is already loaded', async () => {
    vi.mocked(fetchReadReceipts).mockResolvedValueOnce([receipt('alice', 5)])

    await useReadStore.getState().ensureLoaded('chat-1')
    await useReadStore.getState().ensureLoaded('chat-1')

    expect(fetchReadReceipts).toHaveBeenCalledTimes(1)
  })

  it('markRead sets an initial value and advances it', () => {
    useReadStore.getState().markRead('chat-1', 'alice', 5)
    expect(useReadStore.getState().chats['chat-1'].receipts.alice).toBe(5)

    useReadStore.getState().markRead('chat-1', 'alice', 9)
    expect(useReadStore.getState().chats['chat-1'].receipts.alice).toBe(9)
  })

  it('markRead never regresses a receipt (mirrors the backend GREATEST())', () => {
    useReadStore.getState().markRead('chat-1', 'alice', 9)
    useReadStore.getState().markRead('chat-1', 'alice', 3) // stale, out of order

    expect(useReadStore.getState().chats['chat-1'].receipts.alice).toBe(9)
  })

  it('tracks multiple users per chat independently of other chats', () => {
    useReadStore.getState().markRead('chat-1', 'alice', 5)
    useReadStore.getState().markRead('chat-1', 'bob', 2)
    useReadStore.getState().markRead('chat-2', 'carol', 7)

    expect(useReadStore.getState().chats['chat-1'].receipts).toEqual({ alice: 5, bob: 2 })
    expect(useReadStore.getState().chats['chat-2'].receipts).toEqual({ carol: 7 })
  })

  it('refreshAll re-fetches every chat currently tracked', async () => {
    vi.mocked(fetchReadReceipts)
      .mockResolvedValueOnce([receipt('alice', 1)])
      .mockResolvedValueOnce([receipt('bob', 2)])
    await useReadStore.getState().ensureLoaded('chat-1')
    await useReadStore.getState().ensureLoaded('chat-2')
    vi.mocked(fetchReadReceipts).mockClear()

    vi.mocked(fetchReadReceipts)
      .mockResolvedValueOnce([receipt('alice', 10)])
      .mockResolvedValueOnce([receipt('bob', 20)])
    useReadStore.getState().refreshAll()
    await vi.waitFor(() => {
      expect(useReadStore.getState().chats['chat-1'].receipts.alice).toBe(10)
      expect(useReadStore.getState().chats['chat-2'].receipts.bob).toBe(20)
    })
  })

  it('reset clears everything', () => {
    useReadStore.getState().markRead('chat-1', 'alice', 5)

    useReadStore.getState().reset()

    expect(useReadStore.getState().chats).toEqual({})
  })
})
