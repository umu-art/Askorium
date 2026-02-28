import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'

interface MarkdownRendererProps {
  content: string
  className?: string
}

/**
 * Renders markdown content using react-markdown + remark-gfm.
 *
 * Custom component overrides ensure links open in new tabs with
 * proper security attributes and match the brand color palette.
 *
 * Typography is handled by @tailwindcss/typography (`prose` classes).
 */
export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
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
        className
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
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
