export type WsStatus = 'connecting' | 'open' | 'reconnecting' | 'closed'

type MessageListener = (data: string) => void
type StatusListener = (status: WsStatus) => void

const MAX_BACKOFF_MS = 30_000
const BASE_BACKOFF_MS = 1_000

/**
 * Thin wrapper over WebSocket with exponential-backoff reconnect. The caller
 * supplies `getUrl`, called fresh before every connect attempt, so it can
 * hand back a URL with an up-to-date access token.
 */
export class WsClient {
  private socket: WebSocket | null = null
  private readonly messageListeners = new Set<MessageListener>()
  private readonly statusListeners = new Set<StatusListener>()
  private status: WsStatus = 'closed'
  private reconnectAttempt = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private stopped = true
  private readonly getUrl: () => Promise<string>

  constructor(getUrl: () => Promise<string>) {
    this.getUrl = getUrl
  }

  onMessage(listener: MessageListener): () => void {
    this.messageListeners.add(listener)
    return () => this.messageListeners.delete(listener)
  }

  // Fires immediately with the current status, then on every change.
  onStatusChange(listener: StatusListener): () => void {
    this.statusListeners.add(listener)
    listener(this.status)
    return () => this.statusListeners.delete(listener)
  }

  connect(): void {
    this.stopped = false
    void this.open()
  }

  close(): void {
    this.stopped = true
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.socket?.close()
    this.socket = null
    this.setStatus('closed')
  }

  send(data: string): void {
    this.socket?.send(data)
  }

  private setStatus(status: WsStatus): void {
    this.status = status
    for (const listener of this.statusListeners) listener(status)
  }

  private async open(): Promise<void> {
    if (this.stopped) return
    this.setStatus(this.reconnectAttempt > 0 ? 'reconnecting' : 'connecting')

    let url: string
    try {
      url = await this.getUrl()
    } catch {
      this.scheduleReconnect()
      return
    }
    if (this.stopped) return

    const socket = new WebSocket(url)
    this.socket = socket

    socket.addEventListener('open', () => {
      if (this.socket !== socket) return // superseded by a newer attempt
      this.reconnectAttempt = 0
      this.setStatus('open')
    })

    socket.addEventListener('message', (event) => {
      if (this.socket !== socket) return
      for (const listener of this.messageListeners) listener(event.data as string)
    })

    socket.addEventListener('close', () => {
      if (this.socket !== socket) return
      this.socket = null
      if (this.stopped) {
        this.setStatus('closed')
        return
      }
      this.scheduleReconnect()
    })

    // 'close' always follows a failed/dropped connection — nothing extra to
    // do here, just avoid an unhandled-event console warning.
    socket.addEventListener('error', () => {})
  }

  private scheduleReconnect(): void {
    if (this.stopped) return
    this.setStatus('reconnecting')
    const backoff = Math.min(MAX_BACKOFF_MS, BASE_BACKOFF_MS * 2 ** this.reconnectAttempt)
    const jittered = backoff * (0.75 + Math.random() * 0.5)
    this.reconnectAttempt++
    this.reconnectTimer = setTimeout(() => void this.open(), jittered)
  }
}
