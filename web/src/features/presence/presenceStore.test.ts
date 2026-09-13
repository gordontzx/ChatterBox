import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { PresenceInfo } from '../../types'
import { usePresenceStore } from './presenceStore'

vi.mock('./api', () => ({
  fetchPresence: vi.fn(),
}))

import { fetchPresence } from './api'

function presence(online: boolean, lastSeenAt?: string): PresenceInfo {
  return { user_id: 'alice', online, last_seen_at: lastSeenAt }
}

describe('presenceStore', () => {
  beforeEach(() => {
    usePresenceStore.getState().reset()
    vi.clearAllMocks()
  })

  it('ensureLoaded seeds a user as online', async () => {
    vi.mocked(fetchPresence).mockResolvedValueOnce(presence(true))

    await usePresenceStore.getState().ensureLoaded('alice')

    const state = usePresenceStore.getState().byUser.alice
    expect(state.loaded).toBe(true)
    expect(state.online).toBe(true)
  })

  it('ensureLoaded seeds a user as offline with a last-seen time', async () => {
    vi.mocked(fetchPresence).mockResolvedValueOnce(presence(false, '2020-01-01T00:00:00Z'))

    await usePresenceStore.getState().ensureLoaded('alice')

    const state = usePresenceStore.getState().byUser.alice
    expect(state.online).toBe(false)
    expect(state.lastSeenAt).toBe('2020-01-01T00:00:00Z')
  })

  it('does not refetch a user that is already loaded', async () => {
    vi.mocked(fetchPresence).mockResolvedValueOnce(presence(true))

    await usePresenceStore.getState().ensureLoaded('alice')
    await usePresenceStore.getState().ensureLoaded('alice')

    expect(fetchPresence).toHaveBeenCalledTimes(1)
  })

  it('degrades to loaded/offline without throwing when the fetch fails', async () => {
    vi.mocked(fetchPresence).mockRejectedValueOnce(new Error('network down'))

    await usePresenceStore.getState().ensureLoaded('alice')

    const state = usePresenceStore.getState().byUser.alice
    expect(state.loaded).toBe(true)
    expect(state.online).toBe(false)
  })

  it('setPresence applies a live WS update immediately, overwriting prior state', async () => {
    vi.mocked(fetchPresence).mockResolvedValueOnce(presence(false))
    await usePresenceStore.getState().ensureLoaded('alice')

    usePresenceStore.getState().setPresence('alice', true)

    expect(usePresenceStore.getState().byUser.alice.online).toBe(true)
  })

  it('refreshAll re-fetches every user currently tracked', async () => {
    vi.mocked(fetchPresence)
      .mockResolvedValueOnce(presence(false))
      .mockResolvedValueOnce({ user_id: 'bob', online: false })
    await usePresenceStore.getState().ensureLoaded('alice')
    await usePresenceStore.getState().ensureLoaded('bob')
    vi.mocked(fetchPresence).mockClear()

    vi.mocked(fetchPresence)
      .mockResolvedValueOnce(presence(true))
      .mockResolvedValueOnce({ user_id: 'bob', online: true })
    usePresenceStore.getState().refreshAll()
    await vi.waitFor(() => {
      expect(usePresenceStore.getState().byUser.alice.online).toBe(true)
      expect(usePresenceStore.getState().byUser.bob.online).toBe(true)
    })
  })

  it('reset clears everything', () => {
    usePresenceStore.getState().setPresence('alice', true)

    usePresenceStore.getState().reset()

    expect(usePresenceStore.getState().byUser).toEqual({})
  })
})
