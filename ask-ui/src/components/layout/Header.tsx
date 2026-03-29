import { SquarePen, Moon, Sun } from 'lucide-react'
import { LogoIcon } from '@/components/icons/LogoIcon'
import { Button } from '@/components/ui'
import { useTheme } from '@/hooks/useTheme'

interface HeaderProps {
  onNewChat: () => void
}

/**
 * Persistent top navigation bar shown during an active chat session.
 *
 * Kept deliberately minimal — logo + name on the left, theme toggle + "New chat" on the right.
 * This mirrors Claude and Perplexity's header philosophy: the content should dominate,
 * the chrome should disappear.
 */
export function Header({ onNewChat }: HeaderProps) {
  const { theme, toggleTheme } = useTheme()

  return (
    <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b border-gray-100 dark:border-gray-800 bg-white/80 dark:bg-gray-900/80 px-4 backdrop-blur-sm">
      {/* Wordmark */}
      <div className="flex items-center gap-2.5">
        <LogoIcon className="h-7 w-7 rounded-md" />
        <span className="text-base font-semibold text-gray-900 dark:text-gray-100">Askorium</span>
        {/* Subtle HSE scope indicator */}
        <span className="hidden rounded-full border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950 px-2 py-0.5 text-[10px] font-medium text-blue-700 dark:text-blue-400 sm:inline-flex items-center gap-1">
          <span className="block h-1 w-1 rounded-full bg-blue-500 dark:bg-blue-400" aria-hidden="true" />
          НИУ ВШЭ
        </span>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleTheme}
          aria-label={theme === 'dark' ? 'Переключить на светлую тему' : 'Переключить на тёмную тему'}
        >
          {theme === 'dark'
            ? <Sun className="h-4 w-4" />
            : <Moon className="h-4 w-4" />
          }
        </Button>

        <Button
          variant="ghost"
          size="sm"
          onClick={onNewChat}
          aria-label="Начать новый чат"
        >
          <SquarePen className="h-4 w-4" />
          <span className="hidden sm:inline">Новый чат</span>
        </Button>
      </div>
    </header>
  )
}
