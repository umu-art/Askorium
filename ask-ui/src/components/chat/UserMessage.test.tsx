import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { UserMessage } from './UserMessage'

describe('UserMessage', () => {
  it('renders message content', () => {
    render(<UserMessage content="Hello world" />)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })
})
