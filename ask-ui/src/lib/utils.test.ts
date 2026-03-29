import { describe, it, expect } from 'vitest'
import { cn, generateId, extractDomain } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('resolves tailwind conflicts', () => {
    expect(cn('p-2', 'p-4')).toBe('p-4')
  })

  it('handles conditional classes', () => {
    expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz')
  })

  it('handles undefined and null', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })
})

describe('generateId', () => {
  it('returns a non-empty string', () => {
    expect(typeof generateId()).toBe('string')
    expect(generateId().length).toBeGreaterThan(0)
  })

  it('returns unique values', () => {
    const ids = new Set(Array.from({ length: 20 }, generateId))
    expect(ids.size).toBe(20)
  })
})

describe('extractDomain', () => {
  it('extracts hostname from URL', () => {
    expect(extractDomain('https://www.hse.ru/path')).toBe('hse.ru')
  })

  it('strips www prefix', () => {
    expect(extractDomain('https://www.example.com')).toBe('example.com')
  })

  it('keeps subdomain other than www', () => {
    expect(extractDomain('https://docs.example.com/page')).toBe('docs.example.com')
  })

  it('returns original string for invalid URL', () => {
    expect(extractDomain('not-a-url')).toBe('not-a-url')
  })
})
