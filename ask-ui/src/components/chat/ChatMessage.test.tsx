import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ChatMessage } from './ChatMessage'
import type { Message } from '@/types'

vi.mock('@/lib/api', () => ({
  feedbackApi: { submitFeedback: vi.fn() },
  SearchMode: { FAST: 'fast', DEEP: 'deep' },
  SearchStatus: { DONE: 'DONE', FAILED: 'FAILED' },
}))

const base = { id: '1', timestamp: new Date() }

describe('ChatMessage', () => {
  it('renders UserMessage for user role', () => {
    const msg: Message = { ...base, role: 'user', content: 'Hello' }
    render(<ChatMessage message={msg} />)
    expect(screen.getByText('Hello')).toBeInTheDocument()
  })

  it('renders AssistantMessage for assistant role', () => {
    const msg: Message = { ...base, role: 'assistant', content: 'Hi there' }
    render(<ChatMessage message={msg} />)
    expect(screen.getByText('Hi there')).toBeInTheDocument()
  })
})
