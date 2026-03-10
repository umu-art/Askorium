import { useState, useCallback, useEffect, useRef } from 'react'
import type { Message } from '@/types'
import { generateId } from '@/lib/utils'
import { searchApi, sourceApi, SearchStatus } from '@/lib/api'

const POLL_INTERVAL_MS = 500

interface UseChatReturn {
  messages: Message[]
  isLoading: boolean
  inputValue: string
  setInputValue: (value: string) => void
  handleSubmit: (content: string) => Promise<void>
  resetChat: () => void
}

export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const sourceIdRef = useRef<string | null>(null)

  useEffect(() => {
    sourceApi.listSources().then(sources => {
      if (sources.length > 0 && sources[0].id) {
        sourceIdRef.current = sources[0].id
      } else {
        sourceIdRef.current = crypto.randomUUID()
        console.warn('No sources found, using random sourceId:', sourceIdRef.current)
      }
    }).catch(() => {
      sourceIdRef.current = crypto.randomUUID()
      console.warn('Failed to load sources, using random sourceId:', sourceIdRef.current)
    })
  }, [])

  const handleSubmit = useCallback(async (content: string) => {
    if (!content.trim() || isLoading) return

    const userMessage: Message = {
      id: generateId(),
      role: 'user',
      content: content.trim(),
      timestamp: new Date(),
    }

    setMessages(prev => [...prev, userMessage])
    setInputValue('')
    setIsLoading(true)

    try {
      const sourceId = sourceIdRef.current ?? crypto.randomUUID()

      const { queryId } = await searchApi.createSearchQuery({
        searchCreateRequest: { query: content.trim(), sourceId },
      })

      const result = await pollSearchResult(queryId)

      const assistantMessage: Message = {
        id: generateId(),
        role: 'assistant',
        content: result.answer ?? 'Не удалось получить ответ.',
        sources: result.sources,
        timestamp: new Date(),
      }
      setMessages(prev => [...prev, assistantMessage])
    } catch {
      const errorMessage: Message = {
        id: generateId(),
        role: 'assistant',
        content: 'Не удалось получить ответ. Попробуйте ещё раз.',
        timestamp: new Date(),
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }, [isLoading])

  const resetChat = useCallback(() => {
    setMessages([])
    setInputValue('')
    setIsLoading(false)
  }, [])

  return {
    messages,
    isLoading,
    inputValue,
    setInputValue,
    handleSubmit,
    resetChat,
  }
}

async function pollSearchResult(queryId: string) {
  while (true) {
    const result = await searchApi.getSearchQueryResult({ queryId })

    if (result.status === SearchStatus.DONE) return result
    if (result.status === SearchStatus.FAILED) throw new Error('Search failed')

    await new Promise(resolve => setTimeout(resolve, POLL_INTERVAL_MS))
  }
}
