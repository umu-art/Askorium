import { useState, useRef, useCallback } from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { LogoIcon } from '@/components/icons/LogoIcon'
import { MarkdownRenderer } from './MarkdownRenderer'
import { SourceCard } from './SourceCard'
import type { Message } from '@/types'

interface AssistantMessageProps {
  message: Message
}

const INITIAL_COUNT = 3

/**
 * Renders the assistant response: avatar + markdown content + source cards.
 *
 * Sources are displayed in a collapsible grid below the content — first row
 * (up to 3) is always visible, the rest collapse behind a toggle button.
 *
 * Inline citation badges [N] in the text are rendered by MarkdownRenderer;
 * clicking one expands the sources section and highlights the matching card.
 */
export function AssistantMessage({ message }: AssistantMessageProps) {
  const sources = message.sources ?? []
  const [showAll, setShowAll] = useState(false)
  const [highlighted, setHighlighted] = useState<number | null>(null)
  const highlightTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const hasMore = sources.length > INITIAL_COUNT
  const visibleSources = showAll ? sources : sources.slice(0, INITIAL_COUNT)

  const handleCitationClick = useCallback(
    (n: number) => {
      // Expand if the cited source is currently hidden
      if (n > INITIAL_COUNT) setShowAll(true)

      // Pulse-highlight the card
      if (highlightTimer.current) clearTimeout(highlightTimer.current)
      setHighlighted(n)
      highlightTimer.current = setTimeout(() => setHighlighted(null), 1500)

      // Scroll to card after possible state update
      setTimeout(() => {
        document
          .getElementById(`source-${message.id}-${n}`)
          ?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }, 50)
    },
    [message.id],
  )

  return (
    <div className="flex items-start gap-3 animate-slide-up">
      {/* Assistant avatar */}
      <div className="flex-shrink-0 mt-0.5">
        <LogoIcon className="h-7 w-7 rounded-lg" />
      </div>

      <div className="flex-1 min-w-0">
        {/* Markdown answer with inline citation badges */}
        <MarkdownRenderer content={message.content} onCitationClick={handleCitationClick} />

        {/* Source cards */}
        {sources.length > 0 && (
          <div className="mt-5">
            <p className="mb-2 text-[11px] font-medium uppercase tracking-wide text-gray-400">
              Источники
            </p>
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {visibleSources.map((source, i) => (
                <SourceCard
                  key={source.url}
                  source={source}
                  index={i + 1}
                  id={`source-${message.id}-${i + 1}`}
                  highlighted={highlighted === i + 1}
                />
              ))}
            </div>

            {hasMore && (
              <button
                type="button"
                onClick={() => setShowAll((v) => !v)}
                className="mt-2 flex items-center gap-1 text-xs text-gray-400 hover:text-brand-600 transition-colors"
              >
                {showAll ? (
                  <>
                    <ChevronUp className="h-3.5 w-3.5" />
                    Свернуть
                  </>
                ) : (
                  <>
                    <ChevronDown className="h-3.5 w-3.5" />
                    Ещё {sources.length - INITIAL_COUNT}{' '}
                    {pluralSource(sources.length - INITIAL_COUNT)}
                  </>
                )}
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function pluralSource(n: number): string {
  if (n % 10 === 1 && n % 100 !== 11) return 'источник'
  if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) return 'источника'
  return 'источников'
}
