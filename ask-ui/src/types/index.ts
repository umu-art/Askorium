import type { SourceSnippet } from '@/lib/api'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  sources?: SourceSnippet[]
  timestamp: Date
}

export interface ChatState {
  messages: Message[]
  isLoading: boolean
}

export type SendMessageFn = (content: string) => Promise<void>
