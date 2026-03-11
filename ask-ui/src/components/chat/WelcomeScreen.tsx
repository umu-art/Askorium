import { LogoIcon } from '@/components/icons/LogoIcon'
import { ChatInput } from './ChatInput'
import type { SourceDto } from '@/lib/api'
import { extractDomain } from '@/lib/utils'

interface WelcomeScreenProps {
  value: string
  onChange: (value: string) => void
  onSubmit: (value: string) => void
  isLoading: boolean
  sources: SourceDto[]
  selectedSourceId: string | null
  onSourceChange: (id: string) => void
}

const SUGGESTED_QUERIES = [
  'Когда дедлайн сдачи курсовых работ?',
  'Как попасть в общежитие ВШЭ?',
  'Условия академического отпуска',
  'Какие стипендии есть в ВШЭ?',
]

/**
 * Welcome / landing screen — shown before the first message.
 *
 * Layout inspired by Claude and Perplexity:
 * - Logo + product name centered vertically in the viewport
 * - Subtle tagline indicating scope (НИУ ВШЭ)
 * - Main input below, full-width
 * - Suggestion chips below the input to lower the barrier to first interaction
 *
 * The source badge in the wordmark acts as a source selector:
 * - No sources loaded → nothing rendered (loading state)
 * - Single source → static badge showing the domain
 * - Multiple sources → styled <select> for switching between sources
 */
export function WelcomeScreen({ value, onChange, onSubmit, isLoading, sources, selectedSourceId, onSourceChange }: WelcomeScreenProps) {
  const badgeClass = "rounded-full border border-blue-200 bg-blue-50 px-2 py-0.5 text-[10px] font-medium text-blue-700"
  const dotClass = "block h-1 w-1 rounded-full bg-blue-500 shrink-0"

  const sourceBadge = (() => {
    if (sources.length === 0) return null

    if (sources.length === 1) {
      return (
        <span className={`inline-flex items-center gap-1 ${badgeClass}`}>
          <span className={dotClass} aria-hidden="true" />
          {extractDomain(sources[0].sourceUrl ?? '')}
        </span>
      )
    }

    return (
      <span className={`inline-flex items-center gap-1 ${badgeClass}`}>
        <span className={dotClass} aria-hidden="true" />
        <select
          value={selectedSourceId ?? ''}
          onChange={e => onSourceChange(e.target.value)}
          className="appearance-none cursor-pointer bg-transparent text-[10px] font-medium text-blue-700 outline-none"
          disabled={isLoading}
        >
          {sources.map(source => (
            <option key={source.id} value={source.id ?? ''}>
              {extractDomain(source.sourceUrl ?? '')}
            </option>
          ))}
        </select>
      </span>
    )
  })()

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-16">
      {/* Logo + wordmark */}
      <div className="mb-6 flex items-center gap-3">
        <LogoIcon className="h-12 w-12 rounded-lg shadow-sm" />
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-gray-900">Askorium</h1>
          <div className="flex items-center gap-1.5 mt-0.5">
            <p className="text-sm text-gray-500">Умный поиск по сайту</p>
            {sourceBadge}
          </div>
        </div>
      </div>

      {/* Headline */}
      <p className="mb-10 max-w-md text-center text-base text-gray-500 leading-relaxed">
        Задайте вопрос на естественном языке — получите точный ответ&nbsp;со&nbsp;ссылками на источники
      </p>

      {/* Main input */}
      <div className="w-full max-w-2xl">
        <ChatInput
          value={value}
          onChange={onChange}
          onSubmit={onSubmit}
          isLoading={isLoading}
          placeholder="Задайте вопрос о ВШЭ..."
          autoFocus
        />
      </div>

      {/* Suggestion chips */}
      <div className="mt-6 flex flex-wrap justify-center gap-2 max-w-2xl">
        {SUGGESTED_QUERIES.map(query => (
          <button
            key={query}
            onClick={() => onSubmit(query)}
            disabled={isLoading}
            className="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm text-gray-600 shadow-sm transition-all hover:border-brand-300 hover:bg-brand-50 hover:text-brand-700 active:scale-95 disabled:opacity-40"
          >
            {query}
          </button>
        ))}
      </div>
    </div>
  )
}
