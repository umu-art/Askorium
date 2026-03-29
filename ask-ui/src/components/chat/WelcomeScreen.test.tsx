import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WelcomeScreen } from './WelcomeScreen'

vi.mock('@/lib/api', () => ({
  SearchMode: { FAST: 'fast', DEEP: 'deep' },
}))

beforeEach(() => {
  localStorage.clear()
  document.documentElement.classList.remove('dark')
})

const defaultProps = {
  value: '',
  onChange: vi.fn(),
  onSubmit: vi.fn(),
  isLoading: false,
  sources: [],
  selectedSourceId: null,
  onSourceChange: vi.fn(),
  searchMode: 'deep' as const,
  onSearchModeChange: vi.fn(),
}

describe('WelcomeScreen', () => {
  it('renders logo and app name', () => {
    render(<WelcomeScreen {...defaultProps} />)
    expect(screen.getByText('Askorium')).toBeInTheDocument()
  })

  it('renders suggestion chips', () => {
    render(<WelcomeScreen {...defaultProps} />)
    expect(screen.getByText(/дедлайн/i)).toBeInTheDocument()
    expect(screen.getByText(/общежит/i)).toBeInTheDocument()
  })

  it('calls onSubmit when suggestion chip clicked', async () => {
    const onSubmit = vi.fn()
    render(<WelcomeScreen {...defaultProps} onSubmit={onSubmit} />)
    await userEvent.click(screen.getByText(/стипендии/i))
    expect(onSubmit).toHaveBeenCalled()
  })

  it('disables chips when loading', () => {
    render(<WelcomeScreen {...defaultProps} isLoading />)
    const chips = screen.getAllByRole('button').filter(b => b.getAttribute('disabled') !== null)
    expect(chips.length).toBeGreaterThan(0)
  })

  it('renders single source as static badge', () => {
    render(<WelcomeScreen {...defaultProps} sources={[{ id: 's1', sourceUrl: 'https://hse.ru' }]} selectedSourceId="s1" />)
    expect(screen.getByText('hse.ru')).toBeInTheDocument()
  })

  it('renders source selector when multiple sources', () => {
    const sources = [
      { id: 's1', sourceUrl: 'https://hse.ru' },
      { id: 's2', sourceUrl: 'https://example.com' },
    ]
    render(<WelcomeScreen {...defaultProps} sources={sources} selectedSourceId="s1" />)
    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('calls onSourceChange when source selected', async () => {
    const onSourceChange = vi.fn()
    const sources = [
      { id: 's1', sourceUrl: 'https://hse.ru' },
      { id: 's2', sourceUrl: 'https://example.com' },
    ]
    render(<WelcomeScreen {...defaultProps} sources={sources} selectedSourceId="s1" onSourceChange={onSourceChange} />)
    await userEvent.selectOptions(screen.getByRole('combobox'), 's2')
    expect(onSourceChange).toHaveBeenCalledWith('s2')
  })

  it('renders theme toggle button', () => {
    render(<WelcomeScreen {...defaultProps} />)
    expect(screen.getByRole('button', { name: /тём/i })).toBeInTheDocument()
  })
})
