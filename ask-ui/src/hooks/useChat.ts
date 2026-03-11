import { useState, useCallback, useEffect } from 'react'
import type { Message } from '@/types'
import type { SourceDto } from '@/lib/api'
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
  sources: SourceDto[]
  selectedSourceId: string | null
  setSelectedSourceId: (id: string) => void
}

export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const [sources, setSources] = useState<SourceDto[]>([])
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null)

  useEffect(() => {
    sourceApi.listSources().then(list => {
      setSources(list)
      if (list.length > 0 && list[0].id) {
        setSelectedSourceId(list[0].id)
      } else {
        console.warn('No sources found')
      }
    }).catch(() => {
      console.warn('Failed to load sources')
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
      const sourceId = selectedSourceId ?? crypto.randomUUID()

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
  }, [isLoading, selectedSourceId])

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
    sources,
    selectedSourceId,
    setSelectedSourceId,
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
