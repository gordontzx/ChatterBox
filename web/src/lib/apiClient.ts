import { clearTokens, getAccessToken, getRefreshToken, setAccessToken } from './tokens'
import type { AccessTokenResponse } from '../types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// The Go handlers write plain-text bodies via http.Error, not JSON.
async function rawRequest<T>(path: string, init: RequestInit): Promise<T> {
  const res = await fetch(API_BASE_URL + path, init)

  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new ApiError(res.status, text.trim() || res.statusText)
  }
  if (res.status === 204) {
    return undefined as T
  }
  const text = await res.text()
  return (text ? JSON.parse(text) : undefined) as T
}

function authedRequest<T>(path: string, init: RequestInit): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  const token = getAccessToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return rawRequest<T>(path, { ...init, headers })
}

// Only one refresh in flight at a time — concurrent 401s all await the same call.
let refreshPromise: Promise<string> | null = null

// Exported so AuthProvider can call it directly: once on boot (via the 401
// retry below) and periodically to keep the access token from ever expiring
// mid-session.
export function refreshAccessToken(): Promise<string> {
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    const refreshToken = getRefreshToken()
    if (!refreshToken) throw new ApiError(401, 'not logged in')

    const res = await rawRequest<AccessTokenResponse>('/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    setAccessToken(res.access_token)
    return res.access_token
  })()

  return refreshPromise.finally(() => {
    refreshPromise = null
  })
}

/**
 * Authenticated request with one automatic retry-after-refresh on a 401.
 * If there's no refresh token to try (or the refresh itself fails), the
 * original 401 propagates and the caller/AuthProvider treats it as logged out.
 */
export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  try {
    return await authedRequest<T>(path, init)
  } catch (err) {
    if (err instanceof ApiError && err.status === 401 && getRefreshToken()) {
      try {
        await refreshAccessToken()
      } catch {
        clearTokens()
        throw err
      }
      return authedRequest<T>(path, init)
    }
    throw err
  }
}

export const api = {
  get: <T>(path: string) => apiRequest<T>(path, { method: 'GET' }),
  post: <T>(path: string, body?: unknown) =>
    apiRequest<T>(path, {
      method: 'POST',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    }),
}

// GET /chats and GET /chats/{id}/messages serialize an empty result as
// `null`, not `[]` (see internal/chat/store.go, internal/message/store.go).
export function orEmpty<T>(value: T[] | null | undefined): T[] {
  return value ?? []
}
