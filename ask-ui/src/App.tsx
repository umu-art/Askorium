import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ChatPage } from '@/pages/ChatPage'
import { SourcesPage } from '@/pages/SourcesPage'
import { ToastProvider } from '@/context/ToastContext'

function App() {
  return (
    <ToastProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<ChatPage />} />
          <Route path="/sources" element={<SourcesPage />} />
        </Routes>
      </BrowserRouter>
    </ToastProvider>
  )
}

export default App
