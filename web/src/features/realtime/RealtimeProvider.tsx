import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { refreshAccessToken } from '../../lib/apiClient'
import { getAccessToken } from '../../lib/tokens'
import { WsClient, type WsStatus } from '../../lib/ws'
import type { IncomingMessageEvent, OutgoingChatFrame } from '../../types'
import { useAuth } from '../auth/AuthProvider'
import { useMessagesStore } from '../messages/messageStore'

interface RealtimeContextValue {
  status: WsStatus
  sendMessage: (chatId: string, content: string) => void
}

const RealtimeContext = createContext<RealtimeContextValue | null>(null)

function wsUrl(token: string): string {
  const base = import.meta.env.VITE_WS_BASE_URL
  return `${base}/ws?token=${encodeURIComponent(token)}`
}

// Always tries a refresh first so a reconnect (which can happen well after
// the access token's 15 min TTL) dials with a token that's actually valid;
// falls back to whatever's in memory if the refresh call itself fails.
async function getFreshWsUrl(): Promise<string> {
  let token = getAccessToken()
  try {
    token = await refreshAccessToken()
  } catch {
    // use the fallback below
  }
  if (!token) throw new Error('not authenticated')
  return wsUrl(token)
}

export function RealtimeProvider({ children }: { children: ReactNode }) {
  const { status: authStatus } = useAuth()
  const [wsStatus, setWsStatus] = useState<WsStatus>('closed')
  const clientRef = useRef<WsClient | null>(null)

  useEffect(() => {
    if (authStatus !== 'authed') {
      clientRef.current?.close()
      clientRef.current = null
      setWsStatus('closed')
      return
    }

    const client = new WsClient(getFreshWsUrl)
    clientRef.current = client

    const offStatus = client.onStatusChange(setWsStatus)
    const offMessage = client.onMessage((raw) => {
      let parsed: IncomingMessageEvent
      try {
        parsed = JSON.parse(raw) as IncomingMessageEvent
      } catch {
        return // malformed frame — nothing sensible to do with it
      }
      if (parsed.type === 'message') {
        useMessagesStore.getState().upsertMessage(parsed.message.chat_id, parsed.message)
      }
    })

    client.connect()

    return () => {
      offStatus()
      offMessage()
      client.close()
      clientRef.current = null
    }
  }, [authStatus])

  // Backfill anything missed while disconnected, for chats already loaded.
  useEffect(() => {
    if (wsStatus !== 'open') return
    for (const chatId of Object.keys(useMessagesStore.getState().chats)) {
      void useMessagesStore.getState().resync(chatId)
    }
  }, [wsStatus])

  const value = useMemo<RealtimeContextValue>(
    () => ({
      status: wsStatus,
      sendMessage: (chatId, content) => {
        const frame: OutgoingChatFrame = { type: 'message', chat_id: chatId, content }
        clientRef.current?.send(JSON.stringify(frame))
      },
    }),
    [wsStatus],
  )

  return <RealtimeContext.Provider value={value}>{children}</RealtimeContext.Provider>
}

export function useRealtime(): RealtimeContextValue {
  const ctx = useContext(RealtimeContext)
  if (!ctx) throw new Error('useRealtime must be used within a RealtimeProvider')
  return ctx
}
