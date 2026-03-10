import { LogoIcon } from '@/components/icons/LogoIcon'
import { MarkdownRenderer } from './MarkdownRenderer'
import { SourceCard } from './SourceCard'
import type { Message } from '@/types'

interface AssistantMessageProps {
  message: Message
}

/**
 * Renders the assistant response: avatar + markdown content + source cards.
 *
 * Sources are displayed in a responsive grid below the content,
 * matching the Perplexity-style "answer then cite" UX pattern.
 */
export function AssistantMessage({ message }: AssistantMessageProps) {
  const sources = message.sources ?? []

  return (
    <div className="flex items-start gap-3 animate-slide-up">
      {/* Assistant avatar */}
      <div className="flex-shrink-0 mt-0.5">
        <LogoIcon className="h-7 w-7 rounded-lg" />
      </div>

      <div className="flex-1 min-w-0">
        {/* Markdown answer */}
        <MarkdownRenderer content={message.content} />

        {/* Source cards */}
        {sources.length > 0 && (
          <div className="mt-5">
            <p className="mb-2 text-[11px] font-medium uppercase tracking-wide text-gray-400">
              Источники
            </p>
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {sources.map((source, i) => (
                <SourceCard key={source.url} source={source} index={i + 1} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
