// Access token: memory only (never touches disk). Refresh token: localStorage,
// so the session survives a reload. See frontend plan for the XSS tradeoff —
// the backend has no httpOnly-cookie option for this today.

const REFRESH_TOKEN_KEY = 'chatterbox.refresh_token'

let accessToken: string | null = null

export function getAccessToken(): string | null {
  return accessToken
}

export function setAccessToken(token: string | null): void {
  accessToken = token
}

export function getRefreshToken(): string | null {
  try {
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  } catch {
    // Private browsing / storage disabled — treat as logged out.
    return null
  }
}

export function setRefreshToken(token: string | null): void {
  try {
    if (token) {
      localStorage.setItem(REFRESH_TOKEN_KEY, token)
    } else {
      localStorage.removeItem(REFRESH_TOKEN_KEY)
    }
  } catch {
    // Nothing we can do if storage is unavailable; access token still works
    // for the current tab, refresh-on-reload just won't.
  }
}

export function clearTokens(): void {
  accessToken = null
  setRefreshToken(null)
}
