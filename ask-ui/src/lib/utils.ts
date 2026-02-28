import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * Merges Tailwind classes safely, resolving conflicts correctly.
 * Standard pattern from shadcn/ui — preferred over raw clsx.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Generates a random ID for messages.
 * Using crypto.randomUUID when available for proper uniqueness guarantees.
 */
export function generateId(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return Math.random().toString(36).slice(2, 11)
}

/**
 * Extracts the domain from a URL for display purposes.
 */
export function extractDomain(url: string): string {
  try {
    return new URL(url).hostname.replace('www.', '')
  } catch {
    return url
  }
}
