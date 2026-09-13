// Mirrors the JSON shapes served by the Go API. Keep in sync with:
//   internal/auth/types.go, internal/user/types.go,
//   internal/chat/types.go, internal/message/types.go

export interface User {
  id: string
  name: string
  email: string
}

export interface Member {
  id: string
  name: string
}

export type ChatType = 'direct' | 'group'

export interface Chat {
  id: string
  type: ChatType
  name?: string // only present for group chats (omitempty on direct ones)
  created_at: string
  members: Member[]
}

export interface Message {
  id: string
  chat_id: string
  sender_id: string
  seq: number
  content: string
  created_at: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export interface AccessTokenResponse {
  access_token: string
}

// WebSocket frames (internal/message/types.go: IncomingMessage / OutgoingEvent)
export interface OutgoingChatFrame {
  type: 'message'
  chat_id: string
  content: string
}

export interface IncomingMessageEvent {
  type: 'message'
  message: Message
}

export interface OutgoingTypingFrame {
  type: 'typing'
  chat_id: string
}

export interface IncomingTypingEvent {
  type: 'typing'
  chat_id: string
  user_id: string
}

export interface ReadReceipt {
  user_id: string
  last_read_seq: number
}

export interface IncomingReadEvent {
  type: 'read'
  chat_id: string
  user_id: string
  up_to_seq: number
}

export type IncomingEvent = IncomingMessageEvent | IncomingTypingEvent | IncomingReadEvent
