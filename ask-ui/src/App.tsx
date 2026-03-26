import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ChatPage } from '@/pages/ChatPage'
import { SourcesPage } from '@/pages/SourcesPage'
import { ToastProvider } from '@/context/ToastContext'
import { ErrorBoundary } from '@/components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      <ToastProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<ChatPage />} />
            <Route path="/sources" element={<SourcesPage />} />
          </Routes>
        </BrowserRouter>
      </ToastProvider>
    </ErrorBoundary>
  )
}

export default App
