import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'

const { mockListSources, mockUpsertSource, mockDeleteSource, mockSyncSource, mockAutoSyncSource } = vi.hoisted(() => ({
  mockListSources: vi.fn(),
  mockUpsertSource: vi.fn(),
  mockDeleteSource: vi.fn(),
  mockSyncSource: vi.fn(),
  mockAutoSyncSource: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  sourceApi: {
    listSources: mockListSources,
    upsertSource: mockUpsertSource,
    deleteSource: mockDeleteSource,
    syncSource: mockSyncSource,
    autoSyncSource: mockAutoSyncSource,
  },
  SearchMode: { FAST: 'FAST', DEEP: 'DEEP' },
  SearchStatus: { DONE: 'DONE', FAILED: 'FAILED' },
}))

import { useSources } from './useSources'

const sources = [
  { id: 's1', sourceUrl: 'https://hse.ru', syncPolicy: { enabled: false, intervalMinutes: 720 } },
  { id: 's2', sourceUrl: 'https://example.com', syncPolicy: { enabled: true, intervalMinutes: 60 } },
]

beforeEach(() => {
  vi.clearAllMocks()
  mockListSources.mockResolvedValue(sources)
})

describe('useSources', () => {
  it('loads sources on mount', async () => {
    const { result } = renderHook(() => useSources())
    expect(result.current.isLoading).toBe(true)
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.sources).toEqual(sources)
  })

  it('sets error when fetch fails', async () => {
    mockListSources.mockRejectedValueOnce(new Error('Network error'))
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.error).toBe('Не удалось загрузить источники')
  })

  it('upserts a new source', async () => {
    const newSource = { id: 's3', sourceUrl: 'https://new.com', syncPolicy: { enabled: false, intervalMinutes: 720 } }
    mockUpsertSource.mockResolvedValue(newSource)
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    await act(async () => {
      await result.current.upsertSource({ sourceUrl: 'https://new.com', syncPolicy: { enabled: false, intervalMinutes: 720 } })
    })
    expect(result.current.sources).toContainEqual(newSource)
  })

  it('updates existing source in-place', async () => {
    const updated = { ...sources[0], sourceUrl: 'https://hse.ru/updated' }
    mockUpsertSource.mockResolvedValue(updated)
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    await act(async () => { await result.current.upsertSource(updated) })
    expect(result.current.sources.find(s => s.id === 's1')?.sourceUrl).toBe('https://hse.ru/updated')
  })

  it('deletes a source', async () => {
    mockDeleteSource.mockResolvedValue(undefined)
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    await act(async () => { await result.current.deleteSource('s1') })
    expect(result.current.sources.find(s => s.id === 's1')).toBeUndefined()
  })

  it('syncs a source and refreshes the list', async () => {
    mockSyncSource.mockResolvedValue(undefined)
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    await act(async () => { await result.current.syncSource('s1') })
    expect(mockSyncSource).toHaveBeenCalledWith({ sourceSyncRequest: { sourceId: 's1', force: false } })
    expect(mockListSources).toHaveBeenCalledTimes(2)
  })

  it('marks source as syncing during sync', async () => {
    let resolveSyncSource!: () => void
    mockSyncSource.mockReturnValue(new Promise<void>(r => { resolveSyncSource = r }))
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    act(() => { result.current.syncSource('s1') })
    await waitFor(() => expect(result.current.syncingIds.has('s1')).toBe(true))
    await act(async () => resolveSyncSource())
    await waitFor(() => expect(result.current.syncingIds.has('s1')).toBe(false))
  })

  it('triggers auto sync', async () => {
    mockAutoSyncSource.mockResolvedValue(undefined)
    const { result } = renderHook(() => useSources())
    await waitFor(() => expect(result.current.isLoading).toBe(false))
    await act(async () => { await result.current.triggerAutoSync() })
    expect(mockAutoSyncSource).toHaveBeenCalledOnce()
  })
})
