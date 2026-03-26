import { useRef, useEffect, type KeyboardEvent, type FormEvent } from 'react'
import { ArrowUp, Zap, BookOpen } from 'lucide-react'
import { SearchMode } from '@/lib/api'
import { cn } from '@/lib/utils'

interface ChatInputProps {
  value: string
  onChange: (value: string) => void
  onSubmit: (value: string) => void
  isLoading: boolean
  placeholder?: string
  autoFocus?: boolean
  searchMode?: SearchMode
  onSearchModeChange?: (mode: SearchMode) => void
}

/**
 * Auto-resizing textarea input for chat.
 *
 * Design decisions:
 * - Textarea over <input> for multi-line support (key for natural language queries)
 * - Enter submits, Shift+Enter inserts a newline (universal chat convention)
 * - Auto-resize via scrollHeight to avoid scrollbars inside the input
 * - The send button is disabled while empty or loading
 * - Optional search mode toggle (FAST / DEEP) rendered inside the input container
 */
export function ChatInput({
  value,
  onChange,
  onSubmit,
  isLoading,
  placeholder = 'Задайте вопрос...',
  autoFocus = false,
  searchMode,
  onSearchModeChange,
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

  const toggleMode = () => {
    if (!onSearchModeChange || !searchMode) return
    onSearchModeChange(searchMode === SearchMode.DEEP ? SearchMode.FAST : SearchMode.DEEP)
  }

  const isFast = searchMode === SearchMode.FAST

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

        {/* Search mode toggle — only rendered when prop is provided */}
        {searchMode && onSearchModeChange && (
          <button
            type="button"
            onClick={toggleMode}
            aria-label={isFast ? 'Режим: быстрый. Нажмите для глубокого поиска' : 'Режим: глубокий. Нажмите для быстрого поиска'}
            aria-pressed={isFast}
            className={cn(
              'flex-shrink-0 flex items-center gap-1 rounded-lg px-2 h-7 mb-1',
              'text-[11px] font-medium transition-colors',
              isFast
                ? 'bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-800'
                : 'bg-brand-50 dark:bg-brand-900/20 text-brand-600 dark:text-brand-400 border border-brand-200 dark:border-brand-800',
              'hover:opacity-80',
            )}
          >
            {isFast
              ? <><Zap className="h-3 w-3" />Быстрый</>
              : <><BookOpen className="h-3 w-3" />Глубокий</>
            }
          </button>
        )}

        <button
          type="submit"
          disabled={!canSubmit}
          className={cn(
            'flex-shrink-0 flex items-center justify-center',
            'w-8 h-8 rounded-xl mb-1',
            'transition-all duration-150',
            canSubmit
              ? 'bg-brand text-white hover:bg-brand-600 active:bg-brand-700 shadow-sm'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-400 cursor-not-allowed'
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
