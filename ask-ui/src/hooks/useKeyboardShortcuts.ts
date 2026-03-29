import { useEffect } from 'react'

interface Shortcut {
  /** Key to match (case-insensitive). E.g. 'k', 'escape', '/'. */
  key: string
  /** Whether Ctrl (Windows/Linux) or Cmd (Mac) must be held. */
  meta?: boolean
  action: () => void
}

const INPUT_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT'])

/**
 * Registers global keyboard shortcuts via document-level event delegation.
 *
 * Design decisions:
 * - Shortcuts with `meta: false` are suppressed when focus is inside a text
 *   input, preventing interference with normal typing.
 * - Shortcuts with `meta: true` (Cmd/Ctrl combos) fire regardless of focus,
 *   matching the convention of application-level commands.
 * - Cleanup via `removeEventListener` on unmount avoids memory leaks.
 *
 * Usage:
 *   useKeyboardShortcuts([
 *     { key: 'k', meta: true, action: resetChat },
 *   ])
 */
export function useKeyboardShortcuts(shortcuts: Shortcut[]) {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const isMeta = e.metaKey || e.ctrlKey
      const target = e.target as HTMLElement
      const inInput = INPUT_TAGS.has(target.tagName)

      for (const { key, meta = false, action } of shortcuts) {
        if (e.key.toLowerCase() !== key.toLowerCase()) continue
        if (meta !== isMeta) continue
        // Non-meta shortcuts shouldn't fire while typing
        if (!meta && inInput) continue

        e.preventDefault()
        action()
        break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [shortcuts])
}
