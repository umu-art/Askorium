import { createContext, useContext, useState, useCallback, useRef } from 'react'
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react'
import { cn } from '@/lib/utils'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  id: string
  message: string
  type: ToastType
}

interface ToastContextValue {
  toast: {
    success: (message: string) => void
    error: (message: string) => void
    info: (message: string) => void
  }
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

const ToastContext = createContext<ToastContextValue | null>(null)

const TOAST_DURATION_MS = 4000

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>())

  const dismiss = useCallback((id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id))
    const timer = timers.current.get(id)
    if (timer) {
      clearTimeout(timer)
      timers.current.delete(id)
    }
  }, [])

  const add = useCallback((message: string, type: ToastType) => {
    const id = crypto.randomUUID()
    setToasts(prev => [...prev.slice(-4), { id, message, type }]) // max 5 toasts
    timers.current.set(id, setTimeout(() => dismiss(id), TOAST_DURATION_MS))
  }, [dismiss])

  const toast = {
    success: (msg: string) => add(msg, 'success'),
    error:   (msg: string) => add(msg, 'error'),
    info:    (msg: string) => add(msg, 'info'),
  }

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      {/* ARIA assertive — screen readers announce errors immediately */}
      <div
        aria-live="assertive"
        aria-atomic="false"
        className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none"
      >
        {toasts.map(t => (
          <ToastItem key={t.id} toast={t} onDismiss={dismiss} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

// ---------------------------------------------------------------------------
// Toast item
// ---------------------------------------------------------------------------

const ICONS: Record<ToastType, React.ReactNode> = {
  success: <CheckCircle className="h-4 w-4 text-emerald-500 shrink-0" />,
  error:   <AlertCircle className="h-4 w-4 text-red-500 shrink-0" />,
  info:    <Info className="h-4 w-4 text-brand-500 shrink-0" />,
}

const BORDER: Record<ToastType, string> = {
  success: 'border-emerald-100 dark:border-emerald-900',
  error:   'border-red-100 dark:border-red-900',
  info:    'border-brand-100 dark:border-brand-900',
}

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: (id: string) => void }) {
  return (
    <div
      role="alert"
      className={cn(
        'pointer-events-auto flex items-center gap-2.5 rounded-xl px-4 py-3 shadow-lg',
        'border bg-white dark:bg-gray-800 text-sm text-gray-800 dark:text-gray-200',
        'animate-slide-up',
        BORDER[toast.type],
      )}
    >
      {ICONS[toast.type]}
      <span className="flex-1">{toast.message}</span>
      <button
        type="button"
        onClick={() => onDismiss(toast.id)}
        aria-label="Закрыть уведомление"
        className="shrink-0 rounded p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
      >
        <X className="h-3.5 w-3.5" />
      </button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

/**
 * Access toast notifications from any component inside ToastProvider.
 *
 * Usage:
 *   const toast = useToast()
 *   toast.success('Источник синхронизирован')
 *   toast.error('Не удалось загрузить данные')
 */
export function useToast(): ToastContextValue['toast'] {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx.toast
}
