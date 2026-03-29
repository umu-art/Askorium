import { useState, useEffect } from 'react'

type Theme = 'light' | 'dark'

const STORAGE_KEY = 'askorium_theme'

function getInitialTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') return stored
  } catch { /* ignore */ }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

/**
 * Manages light/dark theme.
 *
 * - Reads initial value from localStorage (or system preference on first visit)
 * - Applies `dark` class to <html> for Tailwind's `darkMode: 'class'` strategy
 * - Persists selection to localStorage for subsequent visits
 *
 * The index.html inline script applies the class before React hydrates,
 * eliminating flash of unstyled content (FOUC).
 */
export function useTheme() {
  const [theme, setTheme] = useState<Theme>(getInitialTheme)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
    try { localStorage.setItem(STORAGE_KEY, theme) } catch { /* ignore */ }
  }, [theme])

  const toggleTheme = () => setTheme(t => (t === 'light' ? 'dark' : 'light'))

  return { theme, toggleTheme }
}
