import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'

const { mockCreateSearchQuery, mockGetSearchQueryResult, mockListSources, mockToastError } = vi.hoisted(() => ({
  mockCreateSearchQuery: vi.fn(),
  mockGetSearchQueryResult: vi.fn(),
  mockListSources: vi.fn(),
  mockToastError: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  searchApi: {
    createSearchQuery: mockCreateSearchQuery,
    getSearchQueryResult: mockGetSearchQueryResult,
  },
  sourceApi: { listSources: mockListSources },
  feedbackApi: { submitFeedback: vi.fn() },
  SearchMode: { FAST: 'FAST', DEEP: 'DEEP' },
  SearchStatus: { DONE: 'DONE', FAILED: 'FAILED' },
}))

vi.mock('@/context/ToastContext', () => ({
  useToast: () => ({ success: vi.fn(), error: mockToastError, info: vi.fn() }),
}))

vi.stubEnv('DEV', false)

import { useChat } from './useChat'

const MESSAGES_KEY = 'askorium_chat_messages'
const MODE_KEY = 'askorium_search_mode'

beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  mockListSources.mockResolvedValue([{ id: 'src-1', sourceUrl: 'https://hse.ru' }])
  mockCreateSearchQuery.mockResolvedValue({ queryId: 'q-1' })
  mockGetSearchQueryResult.mockResolvedValue({ status: 'DONE', answer: 'Test answer', sources: [] })
})

describe('useChat', () => {
  it('initialises with empty messages', async () => {
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(mockListSources).toHaveBeenCalled())
    expect(result.current.messages).toHaveLength(0)
    expect(result.current.isLoading).toBe(false)
  })

  it('loads messages from localStorage on mount', async () => {
    const saved = [{ id: '1', role: 'user', content: 'Hi', timestamp: new Date().toISOString() }]
    localStorage.setItem(MESSAGES_KEY, JSON.stringify(saved))
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.messages).toHaveLength(1))
    expect(result.current.messages[0].content).toBe('Hi')
  })

  it('selects first source on load', async () => {
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.selectedSourceId).toBe('src-1'))
  })

  it('handleSubmit adds user and assistant messages', async () => {
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.selectedSourceId).toBe('src-1'))
    await act(async () => { await result.current.handleSubmit('What is HSE?') })
    expect(result.current.messages).toHaveLength(2)
    expect(result.current.messages[0].role).toBe('user')
    expect(result.current.messages[0].content).toBe('What is HSE?')
    expect(result.current.messages[1].role).toBe('assistant')
    expect(result.current.messages[1].content).toBe('Test answer')
  })

  it('ignores empty submit', async () => {
    const { result } = renderHook(() => useChat())
    await act(async () => { await result.current.handleSubmit('   ') })
    expect(result.current.messages).toHaveLength(0)
  })

  it('shows error message when search throws', async () => {
    mockCreateSearchQuery.mockRejectedValueOnce(new Error('Network'))
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.selectedSourceId).toBe('src-1'))
    await act(async () => { await result.current.handleSubmit('Question') })
    expect(mockToastError).toHaveBeenCalled()
    expect(result.current.messages[1].content).toContain('Не удалось')
  })

  it('shows error message when search result is FAILED', async () => {
    mockGetSearchQueryResult.mockResolvedValueOnce({ status: 'FAILED' })
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.selectedSourceId).toBe('src-1'))
    await act(async () => { await result.current.handleSubmit('Question') })
    expect(mockToastError).toHaveBeenCalled()
  })

  it('resetChat clears messages', async () => {
    const { result } = renderHook(() => useChat())
    await waitFor(() => expect(result.current.selectedSourceId).toBe('src-1'))
    await act(async () => { await result.current.handleSubmit('Hi') })
    expect(result.current.messages).toHaveLength(2)
    act(() => result.current.resetChat())
    expect(result.current.messages).toHaveLength(0)
  })

  it('setSearchMode persists to localStorage', async () => {
    const { result } = renderHook(() => useChat())
    act(() => result.current.setSearchMode('FAST' as any))
    expect(localStorage.getItem(MODE_KEY)).toBe('FAST')
    expect(result.current.searchMode).toBe('FAST')
  })

  it('loads searchMode from localStorage', async () => {
    localStorage.setItem(MODE_KEY, 'FAST')
    const { result } = renderHook(() => useChat())
    expect(result.current.searchMode).toBe('FAST')
  })

  it('setInputValue updates inputValue', () => {
    const { result } = renderHook(() => useChat())
    act(() => result.current.setInputValue('hello'))
    expect(result.current.inputValue).toBe('hello')
  })

  it('shows toast when sources fail to load', async () => {
    mockListSources.mockRejectedValueOnce(new Error('fail'))
    renderHook(() => useChat())
    await waitFor(() => expect(mockToastError).toHaveBeenCalledWith('Не удалось загрузить источники'))
  })
})
