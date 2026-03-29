import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Header } from './Header'

beforeEach(() => {
  localStorage.clear()
  document.documentElement.classList.remove('dark')
})

describe('Header', () => {
  it('renders Askorium wordmark', () => {
    render(<Header onNewChat={vi.fn()} />)
    expect(screen.getByText('Askorium')).toBeInTheDocument()
  })

  it('calls onNewChat when new chat button is clicked', async () => {
    const onNewChat = vi.fn()
    render(<Header onNewChat={onNewChat} />)
    await userEvent.click(screen.getByRole('button', { name: /новый чат/i }))
    expect(onNewChat).toHaveBeenCalledOnce()
  })

  it('renders theme toggle button', () => {
    render(<Header onNewChat={vi.fn()} />)
    expect(screen.getByRole('button', { name: /тем/i })).toBeInTheDocument()
  })

  it('toggles theme on click', async () => {
    render(<Header onNewChat={vi.fn()} />)
    const btn = screen.getByRole('button', { name: /тёмную тему/i })
    await userEvent.click(btn)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })
})
