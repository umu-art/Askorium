import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SourcesPage } from './SourcesPage'
import { ToastProvider } from '@/context/ToastContext'

const { mockListSources, mockUpsertSource, mockDeleteSource, mockSyncSource, mockAutoSyncSource } = vi.hoisted(() => ({
  mockListSources: vi.fn(),
  mockUpsertSource: vi.fn(),
  mockDeleteSource: vi.fn(),
  mockSyncSource: vi.fn(),
  mockAutoSyncSource: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  sourceApi: {
    listSources: mockListSources,
    upsertSource: mockUpsertSource,
    deleteSource: mockDeleteSource,
    syncSource: mockSyncSource,
    autoSyncSource: mockAutoSyncSource,
  },
  SearchMode: { FAST: 'FAST', DEEP: 'DEEP' },
  SearchStatus: { DONE: 'DONE', FAILED: 'FAILED' },
}))

const sources = [
  { id: 's1', sourceUrl: 'https://hse.ru', syncPolicy: { enabled: false, intervalMinutes: 720 } },
]

function renderPage() {
  return render(
    <ToastProvider>
      <SourcesPage />
    </ToastProvider>
  )
}

beforeEach(() => {
  vi.resetAllMocks()
  mockListSources.mockResolvedValue(sources)
})

describe('SourcesPage', () => {
  it('shows loading spinner initially', () => {
    mockListSources.mockReturnValue(new Promise(() => {}))
    renderPage()
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('renders source list after load', async () => {
    renderPage()
    await waitFor(() => expect(screen.getByText('hse.ru')).toBeInTheDocument())
  })

  it('shows empty state when no sources', async () => {
    mockListSources.mockResolvedValue([])
    renderPage()
    await waitFor(() => expect(screen.getByText(/нет источников/i)).toBeInTheDocument())
  })

  it('shows error state when fetch fails', async () => {
    mockListSources.mockRejectedValueOnce(new Error('fail'))
    renderPage()
    await waitFor(() => expect(screen.getByText(/не удалось загрузить/i)).toBeInTheDocument())
  })

  it('opens add modal on header "Добавить" button click', async () => {
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    // Header button is the first "Добавить" button
    const addBtns = screen.getAllByRole('button', { name: /добавить/i })
    await userEvent.click(addBtns[0])
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText(/новый источник/i)).toBeInTheDocument()
  })

  it('closes modal on cancel', async () => {
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    await userEvent.click(screen.getAllByRole('button', { name: /добавить/i })[0])
    await userEvent.click(screen.getByRole('button', { name: /отмена/i }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('opens edit modal', async () => {
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    await userEvent.click(screen.getByRole('button', { name: /редактировать/i }))
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText(/редактирование/i)).toBeInTheDocument()
  })

  it('opens delete confirm modal', async () => {
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    await userEvent.click(screen.getByRole('button', { name: /удалить/i }))
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText(/удалить источник\?/i)).toBeInTheDocument()
  })

  it('deletes source on confirm', async () => {
    mockDeleteSource.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    await userEvent.click(screen.getByRole('button', { name: /удалить/i }))
    // Click the "Удалить" button inside the delete confirm dialog
    const dialogDeleteBtn = screen.getByRole('dialog').querySelector('button:last-child')!
    await userEvent.click(dialogDeleteBtn)
    await waitFor(() => expect(mockDeleteSource).toHaveBeenCalledWith({ sourceId: 's1' }))
  })

  it('triggers sync on sync button click', async () => {
    mockSyncSource.mockResolvedValue(undefined)
    renderPage()
    await waitFor(() => screen.getByText('hse.ru'))
    await userEvent.click(screen.getByRole('button', { name: /^синхр\./i }))
    await waitFor(() => expect(mockSyncSource).toHaveBeenCalled())
  })
})
