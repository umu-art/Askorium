import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useKeyboardShortcuts } from './useKeyboardShortcuts'

/**
 * Dispatches a keydown event that bubbles from `target` (default: document).
 * Bubbling is critical: the hook listens at document level, so for input-focus
 * tests the event must originate from the input and propagate upward —
 * only then does e.target.tagName equal 'INPUT'.
 */
function pressKey(
  key: string,
  opts: { ctrlKey?: boolean; metaKey?: boolean; target?: HTMLElement } = {},
) {
  const { target, ...eventInit } = opts
  const event = new KeyboardEvent('keydown', { key, bubbles: true, ...eventInit })
  ;(target ?? document).dispatchEvent(event)
  return event
}

describe('useKeyboardShortcuts', () => {
  it('fires the action when key and modifier both match', () => {
    const action = vi.fn()
    renderHook(() => useKeyboardShortcuts([{ key: 'k', meta: true, action }]))

    pressKey('k', { ctrlKey: true })

    expect(action).toHaveBeenCalledOnce()
  })

  it('is case-insensitive for the key value', () => {
    const action = vi.fn()
    renderHook(() => useKeyboardShortcuts([{ key: 'K', meta: true, action }]))

    pressKey('k', { ctrlKey: true })

    expect(action).toHaveBeenCalledOnce()
  })

  it('does not fire when the modifier does not match (meta required but absent)', () => {
    const action = vi.fn()
    renderHook(() => useKeyboardShortcuts([{ key: 'k', meta: true, action }]))

    pressKey('k') // no Ctrl/Cmd

    expect(action).not.toHaveBeenCalled()
  })

  it('fires meta shortcuts even when focus is inside a text input', () => {
    const action = vi.fn()
    renderHook(() => useKeyboardShortcuts([{ key: 'k', meta: true, action }]))

    const input = document.createElement('input')
    document.body.appendChild(input)
    pressKey('k', { ctrlKey: true, target: input })
    input.remove()

    expect(action).toHaveBeenCalledOnce()
  })

  it('suppresses non-meta shortcuts when focus is inside a text input', () => {
    const action = vi.fn()
    // meta: false (default) — plain key shortcut
    renderHook(() => useKeyboardShortcuts([{ key: 'Escape', action }]))

    const input = document.createElement('input')
    document.body.appendChild(input)
    // Dispatch from the input → e.target.tagName === 'INPUT'
    pressKey('Escape', { target: input })
    input.remove()

    expect(action).not.toHaveBeenCalled()
  })

  it('fires non-meta shortcuts when focus is NOT in an input', () => {
    const action = vi.fn()
    renderHook(() => useKeyboardShortcuts([{ key: 'Escape', action }]))

    pressKey('Escape') // no input focused

    expect(action).toHaveBeenCalledOnce()
  })

  it('removes the event listener on unmount to prevent memory leaks', () => {
    const spy = vi.spyOn(document, 'removeEventListener')
    const { unmount } = renderHook(() =>
      useKeyboardShortcuts([{ key: 'k', meta: true, action: vi.fn() }]),
    )

    unmount()

    expect(spy).toHaveBeenCalledWith('keydown', expect.any(Function))
    spy.mockRestore()
  })

  it('handles multiple shortcuts in a single registration', () => {
    const actionA = vi.fn()
    const actionB = vi.fn()
    renderHook(() =>
      useKeyboardShortcuts([
        { key: 'k', meta: true, action: actionA },
        { key: 'Escape', action: actionB },
      ]),
    )

    pressKey('k', { ctrlKey: true })
    pressKey('Escape')

    expect(actionA).toHaveBeenCalledOnce()
    expect(actionB).toHaveBeenCalledOnce()
  })
})
