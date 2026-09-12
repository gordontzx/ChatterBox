import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Message } from '../../types'
import { PAGE_LIMIT, selectSortedMessages, useMessagesStore } from './messageStore'

vi.mock('./api', () => ({
  fetchMessages: vi.fn(),
}))

import { fetchMessages } from './api'

function msg(seq: number, id = `m${seq}`): Message {
  return {
    id,
    chat_id: 'chat-1',
    sender_id: 'alice',
    seq,
    content: `content ${seq}`,
    created_at: '2020-01-01T00:00:00Z',
  }
}

describe('messagesStore', () => {
  beforeEach(() => {
    useMessagesStore.getState().reset()
    vi.clearAllMocks()
  })

  it('ensureLoaded pages forward until a short read', async () => {
    const page1 = Array.from({ length: PAGE_LIMIT }, (_, i) => msg(i + 1))
    const page2 = [msg(PAGE_LIMIT + 1)]
    vi.mocked(fetchMessages).mockResolvedValueOnce(page1).mockResolvedValueOnce(page2)

    await useMessagesStore.getState().ensureLoaded('chat-1')

    expect(fetchMessages).toHaveBeenCalledTimes(2)
    expect(fetchMessages).toHaveBeenNthCalledWith(1, 'chat-1', 0, PAGE_LIMIT)
    expect(fetchMessages).toHaveBeenNthCalledWith(2, 'chat-1', PAGE_LIMIT, PAGE_LIMIT)

    const state = useMessagesStore.getState().chats['chat-1']
    expect(state.loaded).toBe(true)
    expect(state.lastSeq).toBe(PAGE_LIMIT + 1)
    expect(selectSortedMessages(state)).toHaveLength(PAGE_LIMIT + 1)
  })

  it('does not refetch a chat that is already loaded', async () => {
    vi.mocked(fetchMessages).mockResolvedValueOnce([msg(1)])

    await useMessagesStore.getState().ensureLoaded('chat-1')
    await useMessagesStore.getState().ensureLoaded('chat-1')

    expect(fetchMessages).toHaveBeenCalledTimes(1)
  })

  it('surfaces an error without marking the chat loaded', async () => {
    vi.mocked(fetchMessages).mockRejectedValueOnce(new Error('down'))

    await useMessagesStore.getState().ensureLoaded('chat-1')

    const state = useMessagesStore.getState().chats['chat-1']
    expect(state.loaded).toBe(false)
    expect(state.error).toBeTruthy()
  })

  it('upsertMessage dedupes redelivery by id and keeps lastSeq monotonic', () => {
    useMessagesStore.getState().upsertMessage('chat-1', msg(5))
    useMessagesStore.getState().upsertMessage('chat-1', msg(5)) // redelivered, same id
    useMessagesStore.getState().upsertMessage('chat-1', msg(3))

    const state = useMessagesStore.getState().chats['chat-1']
    expect(selectSortedMessages(state)).toHaveLength(2)
    expect(state.lastSeq).toBe(5)
  })

  it('selectSortedMessages orders by seq regardless of insertion order', () => {
    useMessagesStore.getState().upsertMessage('chat-1', msg(3))
    useMessagesStore.getState().upsertMessage('chat-1', msg(1))
    useMessagesStore.getState().upsertMessage('chat-1', msg(2))

    const seqs = selectSortedMessages(useMessagesStore.getState().chats['chat-1']).map((m) => m.seq)
    expect(seqs).toEqual([1, 2, 3])
  })

  it('resync backfills a loaded chat from its lastSeq and no-ops for one never opened', async () => {
    vi.mocked(fetchMessages).mockResolvedValueOnce([msg(1)])
    await useMessagesStore.getState().ensureLoaded('chat-1')

    vi.mocked(fetchMessages).mockResolvedValueOnce([msg(2)])
    await useMessagesStore.getState().resync('chat-1')

    expect(fetchMessages).toHaveBeenLastCalledWith('chat-1', 1, PAGE_LIMIT)
    expect(selectSortedMessages(useMessagesStore.getState().chats['chat-1'])).toHaveLength(2)

    vi.mocked(fetchMessages).mockClear()
    await useMessagesStore.getState().resync('chat-2') // never loaded
    expect(fetchMessages).not.toHaveBeenCalled()
  })

  // Regression test for the cross-account leak: reset() is what
  // AuthProvider.clearSessionState() relies on to keep one account's
  // message content from surviving into the next account's session.
  it('reset clears everything', () => {
    useMessagesStore.getState().upsertMessage('chat-1', msg(1))

    useMessagesStore.getState().reset()

    expect(useMessagesStore.getState().chats).toEqual({})
  })
})
