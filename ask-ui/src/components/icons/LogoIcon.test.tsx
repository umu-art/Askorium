import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LogoIcon } from './LogoIcon'

describe('LogoIcon', () => {
  it('renders an SVG with accessible label', () => {
    render(<LogoIcon />)
    expect(screen.getByRole('img', { name: 'Askorium logo' })).toBeInTheDocument()
  })

  it('applies className', () => {
    render(<LogoIcon className="h-10 w-10" />)
    expect(screen.getByRole('img')).toHaveClass('h-10', 'w-10')
  })
})
