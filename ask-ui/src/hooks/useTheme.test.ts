import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useTheme } from './useTheme'

const STORAGE_KEY = 'askorium_theme'

beforeEach(() => {
  localStorage.clear()
  document.documentElement.classList.remove('dark')
  // Reset matchMedia to light preference by default
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockReturnValue({ matches: false }),
  })
})

describe('useTheme', () => {
  it('defaults to light theme when no preference stored and system is light', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('light')
  })

  it('defaults to dark theme when system prefers dark', () => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockReturnValue({ matches: true }),
    })
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('dark')
  })

  it('reads initial theme from localStorage', () => {
    localStorage.setItem(STORAGE_KEY, 'dark')
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('dark')
  })

  it('ignores invalid localStorage value', () => {
    localStorage.setItem(STORAGE_KEY, 'invalid')
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('light')
  })

  it('toggles from light to dark', () => {
    const { result } = renderHook(() => useTheme())
    expect(result.current.theme).toBe('light')
    act(() => result.current.toggleTheme())
    expect(result.current.theme).toBe('dark')
  })

  it('toggles from dark to light', () => {
    localStorage.setItem(STORAGE_KEY, 'dark')
    const { result } = renderHook(() => useTheme())
    act(() => result.current.toggleTheme())
    expect(result.current.theme).toBe('light')
  })

  it('applies dark class to documentElement when theme is dark', () => {
    localStorage.setItem(STORAGE_KEY, 'dark')
    renderHook(() => useTheme())
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('removes dark class when theme is light', () => {
    document.documentElement.classList.add('dark')
    localStorage.setItem(STORAGE_KEY, 'light')
    renderHook(() => useTheme())
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('persists theme to localStorage on toggle', () => {
    const { result } = renderHook(() => useTheme())
    act(() => result.current.toggleTheme())
    expect(localStorage.getItem(STORAGE_KEY)).toBe('dark')
  })
})
