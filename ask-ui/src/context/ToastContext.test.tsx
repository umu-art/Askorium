import { describe, it, expect, vi } from 'vitest'
import { render, screen, act, fireEvent } from '@testing-library/react'
import { ToastProvider, useToast } from './ToastContext'

function Page({ type }: { type: 'success' | 'error' | 'info' }) {
  const toast = useToast()
  return <button onClick={() => toast[type](`${type} message`)}>show</button>
}

function wrap(type: 'success' | 'error' | 'info') {
  return render(
    <ToastProvider>
      <Page type={type} />
    </ToastProvider>
  )
}

describe('ToastProvider', () => {
  it('shows success toast', () => {
    wrap('success')
    fireEvent.click(screen.getByText('show'))
    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByText('success message')).toBeInTheDocument()
  })

  it('shows error toast', () => {
    wrap('error')
    fireEvent.click(screen.getByText('show'))
    expect(screen.getByText('error message')).toBeInTheDocument()
  })

  it('shows info toast', () => {
    wrap('info')
    fireEvent.click(screen.getByText('show'))
    expect(screen.getByText('info message')).toBeInTheDocument()
  })

  it('auto-dismisses after 4 seconds', () => {
    vi.useFakeTimers()
    try {
      wrap('success')
      fireEvent.click(screen.getByText('show'))
      expect(screen.getByRole('alert')).toBeInTheDocument()
      act(() => vi.advanceTimersByTime(4000))
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('dismisses on close button click', () => {
    wrap('info')
    fireEvent.click(screen.getByText('show'))
    expect(screen.getByRole('alert')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /закрыть/i }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})

describe('useToast outside provider', () => {
  it('throws when used outside ToastProvider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    function BadComponent() {
      useToast()
      return null
    }
    expect(() => render(<BadComponent />)).toThrow('useToast must be used within ToastProvider')
    spy.mockRestore()
  })
})
