import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Spinner } from './Spinner'

describe('Spinner', () => {
  it('renders with accessible role and label', () => {
    render(<Spinner />)
    expect(screen.getByRole('status', { name: 'Загрузка' })).toBeInTheDocument()
  })

  it.each([
    ['sm', 'h-4'],
    ['md', 'h-6'],
    ['lg', 'h-8'],
  ] as const)('applies %s size class', (size, cls) => {
    render(<Spinner size={size} />)
    expect(screen.getByRole('status')).toHaveClass(cls)
  })

  it('applies custom className', () => {
    render(<Spinner className="custom" />)
    expect(screen.getByRole('status')).toHaveClass('custom')
  })
})
