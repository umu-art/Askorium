import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { MarkdownRenderer } from './MarkdownRenderer'

describe('MarkdownRenderer', () => {
  // ── Markdown rendering ──────────────────────────────────────────────────

  it('renders plain text content', () => {
    render(<MarkdownRenderer content="Hello world" />)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('renders bold markdown', () => {
    render(<MarkdownRenderer content="**bold text**" />)
    expect(screen.getByText('bold text').tagName).toBe('STRONG')
  })

  it('opens links in a new tab with security attributes', () => {
    render(<MarkdownRenderer content="[Visit HSE](https://hse.ru)" />)
    const link = screen.getByRole('link', { name: 'Visit HSE' })
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  // ── Citation rendering ──────────────────────────────────────────────────

  it('renders [N] citations as buttons when onCitationClick is provided', () => {
    render(
      <MarkdownRenderer
        content="First source [1] and second [2]."
        onCitationClick={vi.fn()}
      />,
    )
    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(2)
  })

  it('assigns accessible labels to citation buttons', () => {
    render(
      <MarkdownRenderer content="See [3] for details." onCitationClick={vi.fn()} />,
    )
    expect(
      screen.getByRole('button', { name: 'Перейти к источнику 3' }),
    ).toBeInTheDocument()
  })

  it('calls onCitationClick with the parsed citation number', async () => {
    const user = userEvent.setup()
    const onCitationClick = vi.fn()

    render(
      <MarkdownRenderer content="Reference [5] here." onCitationClick={onCitationClick} />,
    )

    await user.click(screen.getByRole('button', { name: 'Перейти к источнику 5' }))

    expect(onCitationClick).toHaveBeenCalledOnce()
    expect(onCitationClick).toHaveBeenCalledWith(5)
  })

  it('renders no citation buttons when onCitationClick is not provided', () => {
    render(<MarkdownRenderer content="Text with [1] citation" />)
    // [1] should stay as plain text, not become a button
    expect(screen.queryAllByRole('button')).toHaveLength(0)
  })

  it('renders multiple citations in a single paragraph', async () => {
    const user = userEvent.setup()
    const onCitationClick = vi.fn()

    render(
      <MarkdownRenderer
        content="Sources: [1][2][3] support this claim."
        onCitationClick={onCitationClick}
      />,
    )

    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(3)

    await user.click(buttons[2]) // click [3]
    expect(onCitationClick).toHaveBeenCalledWith(3)
  })
})
