import { useState, useCallback } from 'react'
import type { Message } from '@/types'
import { generateId } from '@/lib/utils'
import { sendMessage } from '@/lib/mockApi'

interface UseChatReturn {
  messages: Message[]
  isLoading: boolean
  inputValue: string
  setInputValue: (value: string) => void
  handleSubmit: (content: string) => Promise<void>
  resetChat: () => void
}

/**
 * Core chat state machine.
 *
 * Encapsulates:
 * - Message list management
 * - Loading state
 * - Input value (lifted here so WelcomeScreen and ChatPage share the same input state)
 * - Submit handler (optimistically appends user message, then awaits API)
 * - Reset (for "New chat" button)
 *
 * The API call is abstracted behind sendMessage() in mockApi.ts,
 * making it trivial to swap in real fetch calls later.
 */
export function useChat(): UseChatReturn {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [inputValue, setInputValue] = useState('')

  const handleSubmit = useCallback(async (content: string) => {
    if (!content.trim() || isLoading) return

    // Optimistic: append user message immediately
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
      const assistantMessage = await sendMessage(content.trim())
      setMessages(prev => [...prev, assistantMessage])
    } catch (error) {
      // Append a graceful error message rather than crashing the UI
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
