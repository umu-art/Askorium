import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ChatPage } from '@/pages/ChatPage'
import { SourcesPage } from '@/pages/SourcesPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<ChatPage />} />
        <Route path="/sources" element={<SourcesPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
