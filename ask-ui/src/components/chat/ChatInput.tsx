import { useRef, useEffect, type KeyboardEvent, type FormEvent } from 'react'
import { ArrowUp } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ChatInputProps {
  value: string
  onChange: (value: string) => void
  onSubmit: (value: string) => void
  isLoading: boolean
  placeholder?: string
  autoFocus?: boolean
}

/**
 * Auto-resizing textarea input for chat.
 *
 * Design decisions:
 * - Textarea over <input> for multi-line support (key for natural language queries)
 * - Enter submits, Shift+Enter inserts a newline (universal chat convention)
 * - Auto-resize via scrollHeight to avoid scrollbars inside the input
 * - The send button is disabled while empty or loading
 */
export function ChatInput({
  value,
  onChange,
  onSubmit,
  isLoading,
  placeholder = 'Задайте вопрос...',
  autoFocus = false,
}: ChatInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const canSubmit = value.trim().length > 0 && !isLoading

  // Auto-resize textarea height based on content
  useEffect(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 192)}px` // cap at ~8 lines
  }, [value])

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (canSubmit) handleSubmit()
    }
  }

  const handleSubmit = (e?: FormEvent) => {
    e?.preventDefault()
    if (!canSubmit) return
    onSubmit(value.trim())
  }

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div
        className={cn(
          'relative flex items-end gap-2 bg-white dark:bg-gray-800 rounded-2xl shadow-sm',
          'border border-gray-200 dark:border-gray-700 px-4 py-2',
          'transition-all duration-150',
          'focus-within:border-brand-400 focus-within:ring-2 focus-within:ring-brand-100 dark:focus-within:ring-brand-900'
        )}
      >
        <textarea
          ref={textareaRef}
          value={value}
          onChange={e => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          rows={1}
          placeholder={placeholder}
          autoFocus={autoFocus}
          disabled={isLoading}
          className={cn(
            'flex-1 resize-none bg-transparent outline-none',
            'text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500',
            'py-2 leading-relaxed max-h-48',
            'disabled:opacity-60'
          )}
          aria-label="Поле ввода запроса"
        />

        <button
          type="submit"
          disabled={!canSubmit}
          className={cn(
            'flex-shrink-0 flex items-center justify-center',
            'w-8 h-8 rounded-xl mb-1',
            'transition-all duration-150',
            canSubmit
              ? 'bg-brand text-white hover:bg-brand-600 active:bg-brand-700 shadow-sm'
              : 'bg-gray-100 text-gray-400 cursor-not-allowed'
          )}
          aria-label="Отправить запрос"
        >
          <ArrowUp className="w-4 h-4" strokeWidth={2.5} />
        </button>
      </div>

      <p className="mt-2 text-center text-xs text-gray-400 dark:text-gray-500">
        Enter — отправить &nbsp;·&nbsp; Shift + Enter — новая строка
      </p>
    </form>
  )
}
