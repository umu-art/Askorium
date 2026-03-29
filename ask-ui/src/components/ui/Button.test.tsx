import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { Button } from './Button'

describe('Button', () => {
  // ── Rendering ───────────────────────────────────────────────────────────

  it('renders its children as the accessible name', () => {
    render(<Button>Submit</Button>)
    expect(screen.getByRole('button', { name: 'Submit' })).toBeInTheDocument()
  })

  it('forwards extra props to the underlying <button>', () => {
    render(<Button aria-label="Close dialog" type="submit">×</Button>)
    const btn = screen.getByRole('button')
    expect(btn).toHaveAttribute('type', 'submit')
    expect(btn).toHaveAttribute('aria-label', 'Close dialog')
  })

  // ── Interaction ─────────────────────────────────────────────────────────

  it('calls onClick handler when clicked', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()

    render(<Button onClick={onClick}>Click me</Button>)
    await user.click(screen.getByRole('button'))

    expect(onClick).toHaveBeenCalledOnce()
  })

  // ── Disabled state ──────────────────────────────────────────────────────

  it('is disabled when the disabled prop is set', () => {
    render(<Button disabled>Disabled</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('is disabled when isLoading is true', () => {
    render(<Button isLoading>Save</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('does not call onClick when disabled', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()

    render(<Button disabled onClick={onClick}>Nope</Button>)
    await user.click(screen.getByRole('button'))

    expect(onClick).not.toHaveBeenCalled()
  })

  // ── Loading state ───────────────────────────────────────────────────────

  it('shows a spinner when isLoading is true', () => {
    render(<Button isLoading>Loading…</Button>)
    // Spinner is rendered as a span with the animate-spin Tailwind class
    expect(document.querySelector('.animate-spin')).toBeTruthy()
  })

  it('still shows children text alongside the spinner', () => {
    render(<Button isLoading>Processing</Button>)
    expect(screen.getByText('Processing')).toBeInTheDocument()
  })

  // ── Variants & sizes (smoke) ─────────────────────────────────────────────

  it.each([
    ['primary'],
    ['secondary'],
    ['ghost'],
  ] as const)('renders variant="%s" without errors', (variant) => {
    render(<Button variant={variant}>Label</Button>)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it.each([['sm'], ['md'], ['lg']] as const)(
    'renders size="%s" without errors',
    (size) => {
      render(<Button size={size}>Label</Button>)
      expect(screen.getByRole('button')).toBeInTheDocument()
    },
  )
})
