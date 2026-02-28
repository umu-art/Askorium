import { LogoIcon } from '@/components/icons/LogoIcon'

/**
 * Animated "thinking" indicator shown while the assistant generates a response.
 * Three bouncing dots communicate processing state without text,
 * keeping it language-agnostic and visually clean.
 */
export function ThinkingIndicator() {
  return (
    <div className="flex items-start gap-3 animate-fade-in">
      <div className="flex-shrink-0 mt-0.5">
        <LogoIcon className="h-7 w-7 rounded-lg" />
      </div>
      <div className="flex h-10 items-center gap-1.5 px-1">
        {[0, 150, 300].map(delay => (
          <span
            key={delay}
            className="h-2 w-2 rounded-full bg-brand-400 animate-bounce"
            style={{ animationDelay: `${delay}ms` }}
          />
        ))}
      </div>
    </div>
  )
}
