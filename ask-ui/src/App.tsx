import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/context/ToastContext'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { Spinner } from '@/components/ui'

// ChatPage is the critical path — loaded eagerly (first render, above the fold)
import { ChatPage } from '@/pages/ChatPage'

// SourcesPage is an admin route visited rarely — loaded lazily on first navigation.
// This splits it into a separate JS chunk, reducing the initial bundle size.
const SourcesPage = lazy(() => import('@/pages/SourcesPage').then(m => ({ default: m.SourcesPage })))

function PageLoader() {
  return (
    <div className="flex h-screen items-center justify-center bg-white dark:bg-gray-900">
      <Spinner size="lg" />
    </div>
  )
}

function App() {
  return (
    <ErrorBoundary>
      <ToastProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<ChatPage />} />
            <Route
              path="/sources"
              element={
                <Suspense fallback={<PageLoader />}>
                  <SourcesPage />
                </Suspense>
              }
            />
          </Routes>
        </BrowserRouter>
      </ToastProvider>
    </ErrorBoundary>
  )
}

export default App
