import { useState, useCallback, useEffect } from 'react'
import type { Message } from '@/types'
import { SearchMode, SourceDto, SearchGetResponse } from '@/lib/api'
import { generateId } from '@/lib/utils'
import { searchApi, sourceApi, SearchStatus } from '@/lib/api'
import { useToast } from '@/context/ToastContext'
import { usePolling } from './usePolling'

const SEARCH_MODE_STORAGE_KEY = 'askorium_search_mode'

function loadSearchMode(): SearchMode {
  try {
    const stored = localStorage.getItem(SEARCH_MODE_STORAGE_KEY)
    if (stored === SearchMode.FAST || stored === SearchMode.DEEP) return stored as SearchMode
  } catch { /* ignore */ }
  return SearchMode.DEEP
}

const POLL_INTERVAL_MS = 500
const MESSAGES_STORAGE_KEY = 'askorium_chat_messages'

function loadMessages(): Message[] {
  try {
    const raw = localStorage.getItem(MESSAGES_STORAGE_KEY)
    if (!raw) return []
    // Stored timestamps are ISO strings — convert back to Date objects
    const parsed = JSON.parse(raw) as Array<Omit<Message, 'timestamp'> & { timestamp: string }>
    return parsed.map(m => ({ ...m, timestamp: new Date(m.timestamp) }))
  } catch {
    return []
  }
}

function saveMessages(messages: Message[]) {
  try {
    localStorage.setItem(MESSAGES_STORAGE_KEY, JSON.stringify(messages))
  } catch {
    // Ignore write errors (e.g. storage quota exceeded in private mode)
  }
}

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
  searchMode: SearchMode
  setSearchMode: (mode: SearchMode) => void
}

const DEV_MOCK_MESSAGES: Message[] = import.meta.env.DEV ? [
  { id: 'mock-user', role: 'user', content: 'Какие стипендии есть в ВШЭ?', timestamp: new Date() },
  {
    id: 'mock-assistant',
    role: 'assistant',
    content: `В НИУ ВШЭ действует несколько видов стипендий [1]. Академическая стипендия назначается студентам, успешно сдавшим сессию без задолженностей [2]. Повышенная государственная академическая стипендия (ПГАС) выплачивается за выдающиеся достижения в учёбе, науке или общественной деятельности [3].

Помимо этого, университет выплачивает именные стипендии — например, стипендию Учёного совета ВШЭ и стипендию Правительства РФ [4][5]. Для студентов из малообеспеченных семей предусмотрена государственная социальная стипендия [6].`,
    sources: [
      { title: 'Стипендиальное обеспечение студентов ВШЭ', url: 'https://www.hse.ru/studyspravka/stip/', text: '' },
      { title: 'Академическая стипендия — условия назначения', url: 'https://www.hse.ru/studyspravka/stip/acad/', text: '' },
      { title: 'Повышенная государственная академическая стипендия', url: 'https://www.hse.ru/studyspravka/stip/pgas/', text: '' },
      { title: 'Именные стипендии НИУ ВШЭ', url: 'https://www.hse.ru/studyspravka/stip/named/', text: '' },
      { title: 'Стипендия Правительства Российской Федерации', url: 'https://www.hse.ru/studyspravka/stip/gov/', text: '' },
      { title: 'Социальная стипендия для нуждающихся студентов', url: 'https://www.hse.ru/studyspravka/stip/social/', text: '' },
    ],
    queryId: '00000000-0000-0000-0000-000000000001',
    timestamp: new Date(),
  },
] : []

export function useChat(): UseChatReturn {
  const toast = useToast()
  const { poll } = usePolling<SearchGetResponse>(POLL_INTERVAL_MS)
  const [messages, setMessages] = useState<Message[]>(() => {
    // In DEV mode always use mock data so changes to DEV_MOCK_MESSAGES are visible immediately
    if (import.meta.env.DEV) return DEV_MOCK_MESSAGES
    return loadMessages()
  })
  const [isLoading, setIsLoading] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const [sources, setSources] = useState<SourceDto[]>([])
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(null)
  const [searchMode, setSearchModeState] = useState<SearchMode>(loadSearchMode)

  const setSearchMode = useCallback((mode: SearchMode) => {
    setSearchModeState(mode)
    try { localStorage.setItem(SEARCH_MODE_STORAGE_KEY, mode) } catch { /* ignore */ }
  }, [])

  // Persist messages to localStorage on every change
  useEffect(() => { saveMessages(messages) }, [messages])

  useEffect(() => {
    sourceApi.listSources().then(list => {
      setSources(list)
      if (list.length > 0 && list[0].id) {
        setSelectedSourceId(list[0].id)
      }
    }).catch(() => {
      toast.error('Не удалось загрузить источники')
    })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

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
        searchCreateRequest: { query: content.trim(), mode: searchMode, sourceId },
      })

      const result = await poll(
        () => searchApi.getSearchQueryResult({ queryId }),
        (r) => r.status === SearchStatus.DONE || r.status === SearchStatus.FAILED,
      )
      if (result.status === SearchStatus.FAILED) throw new Error('Search failed')

      const assistantMessage: Message = {
        id: generateId(),
        role: 'assistant',
        content: result.answer ?? 'Не удалось получить ответ.',
        sources: result.sources,
        queryId,
        timestamp: new Date(),
      }
      setMessages(prev => [...prev, assistantMessage])
    } catch {
      toast.error('Не удалось получить ответ. Попробуйте ещё раз.')
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
  }, [isLoading, selectedSourceId, searchMode])

  const resetChat = useCallback(() => {
    setMessages([])
    setInputValue('')
    setIsLoading(false)
    try { localStorage.removeItem(MESSAGES_STORAGE_KEY) } catch { /* ignore */ }
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
    searchMode,
    setSearchMode,
  }
}
