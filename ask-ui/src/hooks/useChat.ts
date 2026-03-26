import { useState, useCallback, useEffect } from 'react'
import type { Message } from '@/types'
import {SearchMode, SourceDto} from '@/lib/api'
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

const DEV_MOCK_MESSAGES: Message[] = import.meta.env.DEV ? [
  { id: 'mock-user', role: 'user', content: 'Какие стипендии есть в ВШЭ?', timestamp: new Date() },
  {
    id: 'mock-assistant',
    role: 'assistant',
    content: `В НИУ ВШЭ действует несколько видов стипендий [1]. Академическая стипендия назначается студентам, успешно сдавшим сессию без задолженностей [2]. Повышенная государственная академическая стипендия (ПГАС) выплачивается за выдающиеся достижения в учёбе, науке или общественной деятельности [3].

Помимо этого, университет выплачивает именные стипендии — например, стипендию Учёного совета ВШЭ и стипендию Правительства РФ [4][5]. Для студентов из малообеспеченных семей предусмотрена государственная социальная стипендия [6].`,
    sources: [
      { title: 'Стипендиальное обеспечение студентов ВШЭ', url: 'https://www.hse.ru/studyspravka/stip/' },
      { title: 'Академическая стипендия — условия назначения', url: 'https://www.hse.ru/studyspravka/stip/acad/' },
      { title: 'Повышенная государственная академическая стипендия', url: 'https://www.hse.ru/studyspravka/stip/pgas/' },
      { title: 'Именные стипендии НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/stip/named/' },
      { title: 'Стипендия Правительства Российской Федерации', url: 'https://www.hse.ru/studyspravka/stip/gov/' },
      { title: 'Социальная стипендия для нуждающихся студентов', url: 'https://www.hse.ru/studyspravka/stip/social/' },
    ],
    timestamp: new Date(),
  },
] : []

export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<Message[]>(DEV_MOCK_MESSAGES)
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
        searchCreateRequest: { query: content.trim(), mode: SearchMode.DEEP, sourceId },
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
