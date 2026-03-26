import type { SourceSnippet } from '@/lib/api'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  sources?: SourceSnippet[]
  queryId?: string   // present on assistant messages — used for feedback
  timestamp: Date
}

export interface ChatState {
  messages: Message[]
  isLoading: boolean
}

export type SendMessageFn = (content: string) => Promise<void>
