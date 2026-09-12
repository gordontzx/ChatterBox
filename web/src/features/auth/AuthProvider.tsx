import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { refreshAccessToken } from '../../lib/apiClient'
import { clearTokens, getRefreshToken, setAccessToken, setRefreshToken } from '../../lib/tokens'
import type { User } from '../../types'
import { fetchMe, login as loginRequest, logout as logoutRequest, registerUser } from './api'

// Keeps the access token from ever expiring mid-session (it's good for 15
// min server-side) so a WS reconnect always has a fresh one to dial with.
const PROACTIVE_REFRESH_INTERVAL_MS = 13 * 60 * 1000

export type AuthStatus = 'booting' | 'anon' | 'authed'

interface AuthContextValue {
  status: AuthStatus
  user: User | null
  login: (email: string, password: string, deviceName?: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  // No refresh token on disk means there's nothing to boot — skip straight
  // to 'anon' instead of a redundant 'booting' render.
  const [status, setStatus] = useState<AuthStatus>(() => (getRefreshToken() ? 'booting' : 'anon'))
  const [user, setUser] = useState<User | null>(null)

  // Boot: if there's a refresh token, fetchMe() will 401 (no access token
  // yet), which apiClient's retry-after-refresh handles transparently.
  useEffect(() => {
    if (!getRefreshToken()) return
    let cancelled = false

    fetchMe()
      .then((me) => {
        if (cancelled) return
        setUser(me)
        setStatus('authed')
      })
      .catch(() => {
        if (cancelled) return
        clearTokens()
        setStatus('anon')
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (status !== 'authed') return

    const id = setInterval(() => {
      refreshAccessToken().catch(() => {
        clearTokens()
        setUser(null)
        setStatus('anon')
      })
    }, PROACTIVE_REFRESH_INTERVAL_MS)

    return () => clearInterval(id)
  }, [status])

  const login = useCallback(async (email: string, password: string, deviceName?: string) => {
    const tokens = await loginRequest({ email, password, deviceName })
    setAccessToken(tokens.access_token)
    setRefreshToken(tokens.refresh_token)
    const me = await fetchMe()
    setUser(me)
    setStatus('authed')
  }, [])

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      await registerUser({ name, email, password })
      await login(email, password)
    },
    [login],
  )

  const logout = useCallback(async () => {
    const refreshToken = getRefreshToken()
    clearTokens()
    setUser(null)
    setStatus('anon')
    if (refreshToken) {
      // Best-effort: local state is already cleared either way.
      await logoutRequest(refreshToken).catch(() => {})
    }
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({ status, user, login, register, logout }),
    [status, user, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}
