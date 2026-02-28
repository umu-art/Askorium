import type { Message } from '@/types'
import { UserMessage } from './UserMessage'
import { AssistantMessage } from './AssistantMessage'

interface ChatMessageProps {
  message: Message
}

/**
 * Dispatcher component: renders the appropriate message variant
 * based on the message role. Keeps ChatPage clean and free of
 * role-based conditionals.
 */
export function ChatMessage({ message }: ChatMessageProps) {
  if (message.role === 'user') {
    return <UserMessage content={message.content} />
  }

  return <AssistantMessage message={message} />
}
