import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useTypingStore } from './typingStore'

describe('typingStore', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    useTypingStore.getState().reset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('marks a user as typing', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')

    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual(['alice'])
  })

  it('does not duplicate a user marked typing twice in a row', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')
    useTypingStore.getState().markTyping('chat-1', 'alice')

    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual(['alice'])
  })

  it('tracks multiple users per chat independently of other chats', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')
    useTypingStore.getState().markTyping('chat-1', 'bob')
    useTypingStore.getState().markTyping('chat-2', 'carol')

    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual(['alice', 'bob'])
    expect(useTypingStore.getState().typingByChat['chat-2']).toEqual(['carol'])
  })

  it('clears a user after the timeout with no further typing events', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')

    vi.advanceTimersByTime(4000)

    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual([])
  })

  it('a repeated typing event resets the timeout instead of stacking', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')
    vi.advanceTimersByTime(3000)
    useTypingStore.getState().markTyping('chat-1', 'alice') // resets the clock

    vi.advanceTimersByTime(3000) // 6s since the first mark, but only 3s since the second
    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual(['alice'])

    vi.advanceTimersByTime(1000) // now 4s since the second mark
    expect(useTypingStore.getState().typingByChat['chat-1']).toEqual([])
  })

  it('reset cancels pending timers and clears everything', () => {
    useTypingStore.getState().markTyping('chat-1', 'alice')

    useTypingStore.getState().reset()
    vi.advanceTimersByTime(10000) // if the timer weren't cancelled, this would throw or resurface state

    expect(useTypingStore.getState().typingByChat).toEqual({})
  })
})
