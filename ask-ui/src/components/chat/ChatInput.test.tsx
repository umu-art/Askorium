import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChatInput } from './ChatInput'

vi.mock('@/lib/api', () => ({
  SearchMode: { FAST: 'fast', DEEP: 'deep' },
}))

function setup(overrides = {}) {
  const props = {
    value: '',
    onChange: vi.fn(),
    onSubmit: vi.fn(),
    isLoading: false,
    ...overrides,
  }
  render(<ChatInput {...props} />)
  return props
}

describe('ChatInput', () => {
  it('renders textarea with placeholder', () => {
    setup({ placeholder: 'Ask something...' })
    expect(screen.getByPlaceholderText('Ask something...')).toBeInTheDocument()
  })

  it('calls onChange when typing', async () => {
    const onChange = vi.fn()
    render(<ChatInput value="" onChange={onChange} onSubmit={vi.fn()} isLoading={false} />)
    await userEvent.type(screen.getByRole('textbox'), 'hi')
    expect(onChange).toHaveBeenCalled()
  })

  it('submit button is disabled when value is empty', () => {
    setup({ value: '' })
    expect(screen.getByRole('button', { name: /отправить/i })).toBeDisabled()
  })

  it('submit button is enabled when value is non-empty', () => {
    setup({ value: 'hello' })
    expect(screen.getByRole('button', { name: /отправить/i })).not.toBeDisabled()
  })

  it('submit button is disabled when loading', () => {
    setup({ value: 'hello', isLoading: true })
    expect(screen.getByRole('button', { name: /отправить/i })).toBeDisabled()
  })

  it('calls onSubmit on form submit', async () => {
    const onSubmit = vi.fn()
    render(<ChatInput value="test query" onChange={vi.fn()} onSubmit={onSubmit} isLoading={false} />)
    await userEvent.click(screen.getByRole('button', { name: /отправить/i }))
    expect(onSubmit).toHaveBeenCalledWith('test query')
  })

  it('calls onSubmit on Enter key', async () => {
    const onSubmit = vi.fn()
    render(<ChatInput value="test" onChange={vi.fn()} onSubmit={onSubmit} isLoading={false} />)
    await userEvent.type(screen.getByRole('textbox'), '{Enter}')
    expect(onSubmit).toHaveBeenCalledWith('test')
  })

  it('does not submit on Shift+Enter', async () => {
    const onSubmit = vi.fn()
    render(<ChatInput value="test" onChange={vi.fn()} onSubmit={onSubmit} isLoading={false} />)
    await userEvent.type(screen.getByRole('textbox'), '{Shift>}{Enter}{/Shift}')
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('renders search mode toggle when provided', () => {
    setup({ searchMode: 'deep', onSearchModeChange: vi.fn() })
    expect(screen.getByRole('button', { name: /режим/i })).toBeInTheDocument()
  })

  it('calls onSearchModeChange when mode toggle clicked', async () => {
    const onSearchModeChange = vi.fn()
    setup({ searchMode: 'deep', onSearchModeChange })
    await userEvent.click(screen.getByRole('button', { name: /режим/i }))
    expect(onSearchModeChange).toHaveBeenCalledWith('fast')
  })

  it('does not render mode toggle when searchMode not provided', () => {
    setup()
    expect(screen.queryByRole('button', { name: /режим/i })).not.toBeInTheDocument()
  })
})
