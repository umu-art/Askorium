import { useState, useEffect, useCallback } from 'react'
import { sourceApi } from '@/lib/api'
import type { SourceDto } from '@/lib/api'

export function useSources() {
  const [sources, setSources] = useState<SourceDto[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [syncingIds, setSyncingIds] = useState<Set<string>>(new Set())
  const [isAutoSyncing, setIsAutoSyncing] = useState(false)

  const fetchSources = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await sourceApi.listSources()
      setSources(data)
    } catch {
      setError('Не удалось загрузить источники')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => { fetchSources() }, [fetchSources])

  const upsertSource = useCallback(async (dto: SourceDto) => {
    const saved = await sourceApi.upsertSource({ sourceDto: dto })
    setSources(prev => {
      const idx = prev.findIndex(s => s.id === saved.id)
      if (idx >= 0) {
        const next = [...prev]
        next[idx] = saved
        return next
      }
      return [...prev, saved]
    })
    return saved
  }, [])

  const deleteSource = useCallback(async (sourceId: string) => {
    await sourceApi.deleteSource({ sourceId })
    setSources(prev => prev.filter(s => s.id !== sourceId))
  }, [])

  const syncSource = useCallback(async (sourceId: string, force = false) => {
    setSyncingIds(prev => new Set(prev).add(sourceId))
    try {
      await sourceApi.syncSource({ sourceSyncRequest: { sourceId, force } })
      await fetchSources()
    } finally {
      setSyncingIds(prev => {
        const next = new Set(prev)
        next.delete(sourceId)
        return next
      })
    }
  }, [fetchSources])

  const triggerAutoSync = useCallback(async () => {
    setIsAutoSyncing(true)
    try {
      await sourceApi.autoSyncSource()
    } finally {
      setIsAutoSyncing(false)
    }
  }, [])

  return {
    sources,
    isLoading,
    error,
    syncingIds,
    isAutoSyncing,
    fetchSources,
    upsertSource,
    deleteSource,
    syncSource,
    triggerAutoSync,
  }
}
