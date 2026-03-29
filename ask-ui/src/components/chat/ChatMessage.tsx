import { memo } from 'react'
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
 *
 * Wrapped in React.memo — the message object is stable after creation
 * (messages array only grows, existing entries never mutate), so this
 * prevents all previously rendered messages from re-rendering when a
 * new message is appended. Without memo: O(n) renders per message.
 * With memo: O(1) — only the new message renders.
 */
export const ChatMessage = memo(function ChatMessage({ message }: ChatMessageProps) {
  if (message.role === 'user') {
    return <UserMessage content={message.content} />
  }

  return <AssistantMessage message={message} />
})
