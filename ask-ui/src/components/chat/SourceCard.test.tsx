import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SourceCard } from './SourceCard'
import type { SourceSnippet } from '@/lib/api'

const source: SourceSnippet = {
  title: 'Стипендии ВШЭ',
  url: 'https://www.hse.ru/stip',
  text: '',
}

describe('SourceCard', () => {
  it('renders source title', () => {
    render(<SourceCard source={source} index={1} />)
    expect(screen.getByText('Стипендии ВШЭ')).toBeInTheDocument()
  })

  it('renders index number', () => {
    render(<SourceCard source={source} index={3} />)
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('renders extracted domain', () => {
    render(<SourceCard source={source} index={1} />)
    expect(screen.getByText('hse.ru')).toBeInTheDocument()
  })

  it('links to source URL with secure attributes', () => {
    render(<SourceCard source={source} index={1} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', source.url)
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('applies highlighted styles when highlighted=true', () => {
    render(<SourceCard source={source} index={1} highlighted />)
    expect(screen.getByRole('link')).toHaveClass('ring-2')
  })

  it('applies custom id', () => {
    render(<SourceCard source={source} index={1} id="source-msg-1" />)
    expect(document.getElementById('source-msg-1')).toBeInTheDocument()
  })
})
