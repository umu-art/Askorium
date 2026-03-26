import React from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'

interface MarkdownRendererProps {
  content: string
  className?: string
  onCitationClick?: (index: number) => void
}

/**
 * Splits a string on [N] citation patterns and returns an array of
 * plain strings and clickable citation badge buttons.
 */
function processCitations(
  children: React.ReactNode,
  onCitationClick: (n: number) => void,
): React.ReactNode {
  return React.Children.map(children, (child) => {
    if (typeof child !== 'string') return child
    const parts = child.split(/(\[\d+\])/)
    if (parts.length <= 1) return child
    return parts.map((part, i) => {
      const m = part.match(/^\[(\d+)\]$/)
      if (!m) return part || null
      const n = parseInt(m[1])
      return (
        <button
          key={i}
          type="button"
          onClick={(e) => {
            e.preventDefault()
            onCitationClick(n)
          }}
          className={cn(
            'inline-flex items-center justify-center',
            'h-[1.6em] min-w-[1.6em] px-[6px]',
            'rounded-md text-[0.72em] font-bold leading-none',
            'bg-brand-100 text-brand-600',
            'hover:bg-brand-500 hover:text-white',
            'transition-colors cursor-pointer mx-px',
            'relative -top-px',
          )}
        >
          {n}
        </button>
      )
    })
  })
}

/**
 * Renders markdown content using react-markdown + remark-gfm.
 *
 * Custom component overrides ensure links open in new tabs with
 * proper security attributes and match the brand color palette.
 *
 * When onCitationClick is provided, inline [N] patterns are rendered
 * as clickable badge buttons instead of plain text.
 *
 * Typography is handled by @tailwindcss/typography (`prose` classes).
 */
export function MarkdownRenderer({ content, className, onCitationClick }: MarkdownRendererProps) {
  const withCitations = (children: React.ReactNode) =>
    onCitationClick ? processCitations(children, onCitationClick) : children

  return (
    <div
      className={cn(
        'prose prose-sm max-w-none',
        'prose-headings:font-semibold prose-headings:text-gray-900',
        'prose-p:text-gray-800 prose-p:leading-relaxed',
        'prose-a:text-brand-500 prose-a:font-medium prose-a:no-underline hover:prose-a:underline',
        'prose-strong:text-gray-900 prose-strong:font-semibold',
        'prose-code:bg-gray-100 prose-code:text-gray-800 prose-code:rounded prose-code:px-1 prose-code:py-0.5 prose-code:text-xs prose-code:font-mono',
        'prose-blockquote:border-l-brand-300 prose-blockquote:text-gray-500 prose-blockquote:not-italic',
        'prose-li:text-gray-800 prose-li:leading-relaxed',
        'prose-ul:my-3 prose-ol:my-3',
        className,
      )}
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          // Ensure all links open in new tab with security attributes
          a: ({ href, children, ...props }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-brand-500 font-medium hover:underline"
              {...props}
            >
              {children}
            </a>
          ),
          // Remove wrapping pre styles — code blocks handled by prose
          pre: ({ children, ...props }) => (
            <pre
              className="bg-gray-50 border border-gray-100 rounded-xl p-4 overflow-x-auto text-xs"
              {...props}
            >
              {children}
            </pre>
          ),
          // Process inline citations inside paragraphs and list items
          p: ({ children, ...props }) => <p {...props}>{withCitations(children)}</p>,
          li: ({ children, ...props }) => <li {...props}>{withCitations(children)}</li>,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
