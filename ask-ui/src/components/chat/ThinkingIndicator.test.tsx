import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ThinkingIndicator } from './ThinkingIndicator'

describe('ThinkingIndicator', () => {
  it('renders accessible status element', () => {
    render(<ThinkingIndicator />)
    expect(screen.getByRole('status', { name: 'Ассистент формирует ответ' })).toBeInTheDocument()
  })
})
