import { Component, type ErrorInfo, type ReactNode } from 'react'
import { AlertCircle, RotateCcw } from 'lucide-react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

/**
 * React Error Boundary — catches render-time exceptions that slip past
 * async error handling (toast notifications cover API errors; this covers
 * component-level crashes).
 *
 * Must be a class component: the React lifecycle methods
 * `getDerivedStateFromError` and `componentDidCatch` are not available
 * as hooks.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // In production this would forward to an error tracking service (e.g. Sentry)
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  handleReset = () => {
    this.setState({ error: null })
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center px-4 bg-white dark:bg-gray-900 animate-fade-in">
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
            <AlertCircle className="h-6 w-6 text-red-500" />
          </div>
          <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">
            Что-то пошло не так
          </h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-6 text-center max-w-xs">
            Произошла непредвиденная ошибка. Попробуйте перезагрузить страницу.
          </p>
          <button
            onClick={this.handleReset}
            className="inline-flex items-center gap-2 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-2.5 text-sm font-medium text-gray-700 dark:text-gray-300 shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            <RotateCcw className="h-4 w-4" />
            Попробовать снова
          </button>
        </div>
      )
    }

    return this.props.children
  }
}
