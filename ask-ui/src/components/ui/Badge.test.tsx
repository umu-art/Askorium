import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from './Badge'

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge>Label</Badge>)
    expect(screen.getByText('Label')).toBeInTheDocument()
  })

  it('uses default variant by default', () => {
    render(<Badge>test</Badge>)
    expect(screen.getByText('test')).toHaveClass('bg-gray-100')
  })

  it.each([
    ['default', 'bg-gray-100'],
    ['brand', 'bg-brand-50'],
    ['muted', 'bg-gray-50'],
  ] as const)('applies %s variant class', (variant, cls) => {
    render(<Badge variant={variant}>x</Badge>)
    expect(screen.getByText('x')).toHaveClass(cls)
  })

  it('forwards className and HTML props', () => {
    render(<Badge className="extra" data-testid="badge">x</Badge>)
    const el = screen.getByTestId('badge')
    expect(el).toHaveClass('extra')
    expect(el.tagName).toBe('SPAN')
  })
})
