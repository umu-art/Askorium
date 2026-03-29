import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AssistantMessage } from './AssistantMessage'
import type { Message } from '@/types'

const { mockSubmitFeedback } = vi.hoisted(() => ({
  mockSubmitFeedback: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('@/lib/api', () => ({
  feedbackApi: { submitFeedback: mockSubmitFeedback },
  SearchMode: { FAST: 'FAST', DEEP: 'DEEP' },
}))

const baseMessage: Message = {
  id: 'msg-1',
  role: 'assistant',
  content: 'Hello world',
  timestamp: new Date(),
}

const messageWithSources: Message = {
  ...baseMessage,
  queryId: 'q-1',
  sources: [
    { title: 'Source 1', url: 'https://example.com/1', text: '' },
    { title: 'Source 2', url: 'https://example.com/2', text: '' },
    { title: 'Source 3', url: 'https://example.com/3', text: '' },
    { title: 'Source 4', url: 'https://example.com/4', text: '' },
  ],
}

beforeEach(() => mockSubmitFeedback.mockClear())

describe('AssistantMessage', () => {
  it('renders message content', () => {
    render(<AssistantMessage message={baseMessage} />)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('renders copy button', () => {
    render(<AssistantMessage message={baseMessage} />)
    expect(screen.getByRole('button', { name: /скопировать/i })).toBeInTheDocument()
  })

  it('renders feedback buttons when queryId present', () => {
    render(<AssistantMessage message={{ ...baseMessage, queryId: 'q-1' }} />)
    expect(screen.getByRole('button', { name: 'Ответ полезен' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Ответ не полезен' })).toBeInTheDocument()
  })

  it('does not render feedback buttons without queryId', () => {
    render(<AssistantMessage message={baseMessage} />)
    expect(screen.queryByRole('button', { name: /полезен/i })).not.toBeInTheDocument()
  })

  it('renders sources section when sources provided', () => {
    render(<AssistantMessage message={messageWithSources} />)
    expect(screen.getByText('Источники')).toBeInTheDocument()
  })

  it('shows first 3 sources and expand button', () => {
    render(<AssistantMessage message={messageWithSources} />)
    expect(screen.getByText('Source 1')).toBeInTheDocument()
    expect(screen.getByText('Source 3')).toBeInTheDocument()
    expect(screen.queryByText('Source 4')).not.toBeInTheDocument()
    expect(screen.getByText(/ещё/i)).toBeInTheDocument()
  })

  it('expands to show all sources on toggle', async () => {
    render(<AssistantMessage message={messageWithSources} />)
    await userEvent.click(screen.getByText(/ещё/i))
    expect(screen.getByText('Source 4')).toBeInTheDocument()
    expect(screen.getByText(/свернуть/i)).toBeInTheDocument()
  })

  it('collapses sources on second toggle', async () => {
    render(<AssistantMessage message={messageWithSources} />)
    await userEvent.click(screen.getByText(/ещё/i))
    await userEvent.click(screen.getByText(/свернуть/i))
    expect(screen.queryByText('Source 4')).not.toBeInTheDocument()
  })

  it('submits positive feedback', async () => {
    render(<AssistantMessage message={{ ...baseMessage, queryId: 'q-1' }} />)
    await userEvent.click(screen.getByRole('button', { name: /^ответ полезен/i }))
    expect(mockSubmitFeedback).toHaveBeenCalledWith({
      feedbackDto: { queryId: 'q-1', rating: 5 },
    })
  })

  it('submits negative feedback', async () => {
    render(<AssistantMessage message={{ ...baseMessage, queryId: 'q-1' }} />)
    await userEvent.click(screen.getByRole('button', { name: /не полезен/i }))
    expect(mockSubmitFeedback).toHaveBeenCalledWith({
      feedbackDto: { queryId: 'q-1', rating: 1 },
    })
  })

  it('disables feedback buttons after submission', async () => {
    render(<AssistantMessage message={{ ...baseMessage, queryId: 'q-1' }} />)
    await userEvent.click(screen.getByRole('button', { name: /^ответ полезен/i }))
    expect(screen.getByRole('button', { name: /^ответ полезен/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: /не полезен/i })).toBeDisabled()
  })
})
