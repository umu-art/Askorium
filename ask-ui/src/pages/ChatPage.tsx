import { useEffect, useRef, useMemo } from 'react'
import { Header } from '@/components/layout/Header'
import { ChatMessage } from '@/components/chat/ChatMessage'
import { ChatInput } from '@/components/chat/ChatInput'
import { ThinkingIndicator } from '@/components/chat/ThinkingIndicator'
import { WelcomeScreen } from '@/components/chat/WelcomeScreen'
import { useChat } from '@/hooks/useChat'
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts'

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
  const { messages, isLoading, inputValue, setInputValue, handleSubmit, resetChat, sources, selectedSourceId, setSelectedSourceId, searchMode, setSearchMode } = useChat()
  const bottomRef = useRef<HTMLDivElement>(null)
  const hasMessages = messages.length > 0

  // Auto-scroll to bottom on new messages / loading state change
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isLoading])

  // Ctrl/Cmd+K → new chat (mirrors VS Code / Perplexity convention)
  const shortcuts = useMemo(() => [
    { key: 'k', meta: true, action: resetChat },
  ], [resetChat])
  useKeyboardShortcuts(shortcuts)

  if (!hasMessages) {
    return (
      <WelcomeScreen
        value={inputValue}
        onChange={setInputValue}
        onSubmit={handleSubmit}
        isLoading={isLoading}
        sources={sources}
        selectedSourceId={selectedSourceId}
        onSourceChange={setSelectedSourceId}
        searchMode={searchMode}
        onSearchModeChange={setSearchMode}
      />
    )
  }

  return (
    <div className="flex h-screen flex-col bg-white dark:bg-gray-900">
      <Header onNewChat={resetChat} />

      {/* Scrollable messages area */}
      <main className="flex-1 overflow-y-auto scrollbar-gutter-stable">
        <div
          className="mx-auto max-w-3xl px-4 py-8 space-y-8"
          role="log"
          aria-live="polite"
          aria-label="История переписки"
        >
          {messages.map(message => (
            <ChatMessage key={message.id} message={message} />
          ))}

          {isLoading && <ThinkingIndicator />}

          {/* Invisible anchor for auto-scroll */}
          <div ref={bottomRef} />
        </div>
      </main>

      {/* Sticky input at bottom */}
      <div className="border-t border-gray-100 dark:border-gray-800 bg-white/90 dark:bg-gray-900/90 backdrop-blur-sm">
        <div className="mx-auto max-w-3xl px-4 py-3">
          <ChatInput
            value={inputValue}
            onChange={setInputValue}
            onSubmit={handleSubmit}
            isLoading={isLoading}
            searchMode={searchMode}
            onSearchModeChange={setSearchMode}
          />
        </div>
      </div>
    </div>
  )
}
