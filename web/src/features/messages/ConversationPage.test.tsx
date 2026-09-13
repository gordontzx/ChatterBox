// @vitest-environment jsdom
import { act } from 'react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Chat, Message, User } from '../../types'
import { useChatsStore } from '../chats/chatStore'
import { ConversationPage } from './ConversationPage'
import { useMessagesStore } from './messageStore'
import { useReadStore } from './readStore'

const SELF: User = { id: 'self', name: 'Self', email: 'self@example.com' }

vi.mock('../auth/AuthProvider', () => ({
  useAuth: () => ({
    status: 'authed',
    user: SELF,
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  }),
}))

vi.mock('../realtime/RealtimeProvider', () => ({
  useRealtime: () => ({ status: 'open', sendMessage: vi.fn(), sendTyping: vi.fn() }),
}))

vi.mock('./api', () => ({
  fetchMessages: vi.fn().mockResolvedValue([]),
  markAsRead: vi.fn().mockResolvedValue(undefined),
  fetchReadReceipts: vi.fn().mockResolvedValue([]),
}))

function chat(id: string, otherName: string): Chat {
  return {
    id,
    created_at: '2020-01-01T00:00:00Z',
    members: [
      { id: 'self', name: 'Self' },
      { id: `${id}-other`, name: otherName },
    ],
  }
}

function message(chatId: string, id: string, content: string): Message {
  return {
    id,
    chat_id: chatId,
    sender_id: 'self',
    seq: 1,
    content,
    created_at: '2020-01-01T00:00:00Z',
  }
}

describe('ConversationPage', () => {
  // This project's vitest config doesn't set `test.globals`, so
  // @testing-library/react's automatic afterEach(cleanup) never registers
  // — without this, a previous test's tree stays mounted and subscribed to
  // the (shared, module-level) zustand stores, so a later test seeding
  // those stores makes the stale tree re-render too, corrupting queries
  // that expect exactly one match.
  afterEach(() => {
    cleanup()
  })

  beforeEach(() => {
    // jsdom doesn't implement scrollIntoView.
    window.HTMLElement.prototype.scrollIntoView = vi.fn()

    useChatsStore.getState().reset()
    useMessagesStore.getState().reset()
    useReadStore.getState().reset()
    useChatsStore.setState({
      chats: [chat('bob-chat', 'Bob'), chat('charlie-chat', 'Charlie')],
      loaded: true,
    })
    useMessagesStore.setState({
      chats: {
        'bob-chat': {
          byId: { m1: message('bob-chat', 'm1', 'hi bob') },
          lastSeq: 1,
          loading: false,
          loaded: true,
          error: null,
        },
        'charlie-chat': {
          byId: { m2: message('charlie-chat', 'm2', 'hi charlie') },
          lastSeq: 1,
          loading: false,
          loaded: true,
          error: null,
        },
      },
    })
  })

  it('does not accumulate MessageList instances when toggling between two chats', async () => {
    const router = createMemoryRouter([{ path: '/chats/:chatId', element: <ConversationPage /> }], {
      initialEntries: ['/chats/bob-chat'],
    })
    render(<RouterProvider router={router} />)

    for (let i = 0; i < 6; i++) {
      const target = i % 2 === 0 ? '/chats/charlie-chat' : '/chats/bob-chat'
      // eslint-disable-next-line no-await-in-loop
      await act(async () => {
        await router.navigate(target)
      })
    }

    expect(screen.getAllByTestId('message-list')).toHaveLength(1)
  })

  it('shows the seen tick on my own message once the other member has read up to its seq', async () => {
    useReadStore.setState({
      chats: { 'bob-chat': { receipts: { 'bob-chat-other': 1 }, loaded: true, loading: false } },
    })
    const router = createMemoryRouter([{ path: '/chats/:chatId', element: <ConversationPage /> }], {
      initialEntries: ['/chats/bob-chat'],
    })

    render(<RouterProvider router={router} />)

    expect(await screen.findByText('✓✓')).toBeTruthy()
  })

  it('does not show the seen tick when the other member has not caught up yet', async () => {
    useReadStore.setState({
      chats: { 'bob-chat': { receipts: { 'bob-chat-other': 0 }, loaded: true, loading: false } },
    })
    const router = createMemoryRouter([{ path: '/chats/:chatId', element: <ConversationPage /> }], {
      initialEntries: ['/chats/bob-chat'],
    })

    render(<RouterProvider router={router} />)

    expect(screen.queryByText('✓✓')).toBeNull()
  })
})
