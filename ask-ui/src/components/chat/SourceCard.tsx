import { ExternalLink } from 'lucide-react'
import type { SourceSnippet } from '@/lib/api'
import { cn, extractDomain } from '@/lib/utils'

interface SourceCardProps {
  source: SourceSnippet
  index: number
  className?: string
  highlighted?: boolean
  id?: string
}

/**
 * Compact card linking to a source document.
 *
 * Design: numbered badge + title + domain — matches the Perplexity pattern
 * where citations are numbered inline in the text and the cards are listed below.
 * Cards are interactive (hover state) to signal they are clickable.
 */
export function SourceCard({ source, index, className, highlighted, id }: SourceCardProps) {
  return (
    <a
      id={id}
      href={source.url}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        'group flex items-start gap-2.5 rounded-xl border bg-gray-50 p-3',
        'transition-all duration-200',
        highlighted
          ? 'border-brand-400 bg-brand-50 ring-2 ring-brand-200'
          : 'border-gray-100 hover:border-brand-200 hover:bg-brand-50',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-400',
        className
      )}
      aria-label={`Источник ${index}: ${source.title}`}
    >
      {/* Citation number */}
      <span className="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-md bg-gray-200 text-[10px] font-bold text-gray-500 transition-colors group-hover:bg-brand-100 group-hover:text-brand-600">
        {index}
      </span>

      {/* Title + domain */}
      <div className="min-w-0 flex-1">
        <p className="line-clamp-2 text-xs font-medium leading-snug text-gray-700 transition-colors group-hover:text-brand-700">
          {source.title}
        </p>
        <p className="mt-0.5 truncate text-[10px] text-gray-400">{extractDomain(source.url)}</p>
      </div>

      {/* External link icon */}
      <ExternalLink className="mt-0.5 h-3 w-3 flex-shrink-0 text-gray-300 transition-colors group-hover:text-brand-400" />
    </a>
  )
}
