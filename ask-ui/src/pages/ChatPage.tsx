import { useEffect, useRef } from 'react'
import { Header } from '@/components/layout/Header'
import { ChatMessage } from '@/components/chat/ChatMessage'
import { ChatInput } from '@/components/chat/ChatInput'
import { ThinkingIndicator } from '@/components/chat/ThinkingIndicator'
import { WelcomeScreen } from '@/components/chat/WelcomeScreen'
import { useChat } from '@/hooks/useChat'

/**
 * Main page component — the single page of the application.
 *
 * State machine:
 * - No messages → WelcomeScreen (full-viewport centered layout)
 * - Messages present → Chat layout (header + scrollable messages + sticky input)
 *
 * The transition between states is implicit: once the first message is sent,
 * `messages.length > 0` triggers the chat layout to render.
 *
 * Auto-scroll: the bottomRef div at the end of the message list is scrolled into
 * view whenever messages or isLoading change. Using `behavior: 'smooth'` for UX polish.
 */
export function ChatPage() {
  const { messages, isLoading, inputValue, setInputValue, handleSubmit, resetChat } = useChat()
  const bottomRef = useRef<HTMLDivElement>(null)
  const hasMessages = messages.length > 0

  // Auto-scroll to bottom on new messages / loading state change
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isLoading])

  if (!hasMessages) {
    return (
      <WelcomeScreen
        value={inputValue}
        onChange={setInputValue}
        onSubmit={handleSubmit}
        isLoading={isLoading}
      />
    )
  }

  return (
    <div className="flex h-screen flex-col bg-white">
      <Header onNewChat={resetChat} />

      {/* Scrollable messages area */}
      <main className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl px-4 py-8 space-y-8">
          {messages.map(message => (
            <ChatMessage key={message.id} message={message} />
          ))}

          {isLoading && <ThinkingIndicator />}

          {/* Invisible anchor for auto-scroll */}
          <div ref={bottomRef} />
        </div>
      </main>

      {/* Sticky input at bottom */}
      <div className="border-t border-gray-100 bg-white/90 backdrop-blur-sm">
        <div className="mx-auto max-w-3xl px-4 py-3">
          <ChatInput
            value={inputValue}
            onChange={setInputValue}
            onSubmit={handleSubmit}
            isLoading={isLoading}
          />
        </div>
      </div>
    </div>
  )
}
