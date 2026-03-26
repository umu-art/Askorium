interface UserMessageProps {
  content: string
}

/**
 * Renders the user's message.
 *
 * Deliberately simple — right-aligned pill.
 * This asymmetry makes user vs assistant messages instantly distinguishable
 * (same pattern as iMessage, Claude, Perplexity).
 */
export function UserMessage({ content }: UserMessageProps) {
  return (
    <div className="flex justify-end">
      <div className="max-w-[75%] rounded-2xl rounded-br-sm bg-brand-50 dark:bg-brand-900/30 px-4 py-3 text-sm leading-relaxed text-gray-900 dark:text-gray-100 shadow-sm border border-brand-100 dark:border-brand-800">
        {content}
      </div>
    </div>
  )
}
