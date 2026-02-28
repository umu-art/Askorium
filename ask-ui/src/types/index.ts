export interface Source {
  id: number
  title: string
  url: string
  domain: string
  snippet?: string
}

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  sources?: Source[]
  timestamp: Date
}

export interface ChatState {
  messages: Message[]
  isLoading: boolean
}

export type SendMessageFn = (content: string) => Promise<void>
