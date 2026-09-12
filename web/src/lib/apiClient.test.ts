import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { apiRequest, orEmpty } from './apiClient'
import { clearTokens, getAccessToken, setAccessToken, setRefreshToken } from './tokens'

function jsonResponse(body: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    text: async () => JSON.stringify(body),
  } as unknown as Response
}

function textErrorResponse(status: number, message: string) {
  return {
    ok: false,
    status,
    text: async () => message,
  } as unknown as Response
}

describe('apiRequest', () => {
  beforeEach(() => {
    clearTokens()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('attaches the bearer token when one is set', async () => {
    setAccessToken('token-123')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(jsonResponse({ ok: true }))

    await apiRequest('/whatever')

    const init = fetchMock.mock.calls[0][1]
    const headers = init?.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer token-123')
  })

  it('retries once with a fresh token after a successful refresh on 401', async () => {
    setRefreshToken('refresh-abc')
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(textErrorResponse(401, 'unauthorized')) // original request
      .mockResolvedValueOnce(jsonResponse({ access_token: 'new-token' })) // /auth/refresh
      .mockResolvedValueOnce(jsonResponse({ hello: 'world' })) // retried request

    const result = await apiRequest('/protected')

    expect(result).toEqual({ hello: 'world' })
    expect(getAccessToken()).toBe('new-token')
    expect(fetchMock).toHaveBeenCalledTimes(3)
  })

  it('propagates the original 401 when there is no refresh token to try', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(textErrorResponse(401, 'nope'))

    await expect(apiRequest('/protected')).rejects.toMatchObject({ status: 401 })
  })

  it('propagates the original 401 and clears tokens when the refresh itself fails', async () => {
    setAccessToken('stale-token')
    setRefreshToken('refresh-abc')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(textErrorResponse(401, 'expired'))
      .mockResolvedValueOnce(textErrorResponse(401, 'refresh invalid'))

    await expect(apiRequest('/protected')).rejects.toMatchObject({ status: 401 })
    expect(getAccessToken()).toBeNull()
  })

  it('coalesces concurrent refreshes into a single request', async () => {
    setRefreshToken('refresh-abc')
    let refreshCalls = 0
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const path = String(input)
      if (path.endsWith('/auth/refresh')) {
        refreshCalls++
        return jsonResponse({ access_token: 'new-token' })
      }
      if (getAccessToken() === 'new-token') return jsonResponse({ ok: true })
      return textErrorResponse(401, 'unauthorized')
    })

    await Promise.all([apiRequest('/a'), apiRequest('/b')])

    expect(refreshCalls).toBe(1)
  })
})

describe('orEmpty', () => {
  it('coerces null/undefined to an empty array', () => {
    expect(orEmpty(null)).toEqual([])
    expect(orEmpty(undefined)).toEqual([])
  })

  it('passes arrays through unchanged', () => {
    expect(orEmpty([1, 2, 3])).toEqual([1, 2, 3])
  })
})
