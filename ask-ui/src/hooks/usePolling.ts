import { useRef, useCallback, useEffect } from 'react'

/**
 * Generic polling hook with AbortController-based cancellation.
 *
 * Usage:
 *   const { poll } = usePolling<ResultDto>(500)
 *   const result = await poll(
 *     () => api.getResult({ id }),
 *     (r) => r.status !== 'IN_PROGRESS',
 *   )
 *
 * Design decisions:
 * - `poll()` aborts any in-flight polling before starting a new one —
 *   safe to call multiple times without leaking intervals.
 * - Unmount cleanup: the current AbortController is aborted when the
 *   component that owns this hook unmounts, preventing state updates
 *   on unmounted components.
 * - AbortError is re-thrown so callers can distinguish cancellation from
 *   actual errors (catch AbortError separately if needed).
 * - The interval delay itself is also abortable — no dangling timers.
 */
export function usePolling<T>(intervalMs = 500) {
  const abortRef = useRef<AbortController | null>(null)

  // Cancel any in-flight polling on unmount
  useEffect(() => () => { abortRef.current?.abort() }, [])

  const poll = useCallback(
    async (
      fetcher: () => Promise<T>,
      isDone: (result: T) => boolean,
    ): Promise<T> => {
      // Abort any previous poll before starting a new one
      abortRef.current?.abort()
      abortRef.current = new AbortController()
      const { signal } = abortRef.current

      while (!signal.aborted) {
        const result = await fetcher()
        if (isDone(result)) return result

        // Interruptible sleep: rejects immediately when aborted.
        // We also check signal.aborted synchronously at construction time —
        // the abort may have occurred between `await fetcher()` returning and
        // this point (e.g. when a second poll() call cancels the first one
        // before the abort listener has been registered).
        await new Promise<void>((resolve, reject) => {
          if (signal.aborted) {
            reject(new DOMException('Polling aborted', 'AbortError'))
            return
          }
          const timer = setTimeout(resolve, intervalMs)
          signal.addEventListener('abort', () => {
            clearTimeout(timer)
            reject(new DOMException('Polling aborted', 'AbortError'))
          }, { once: true })
        })
      }

      // Reached only when aborted before fetcher resolves
      throw new DOMException('Polling aborted', 'AbortError')
    },
    [intervalMs],
  )

  const cancel = useCallback(() => { abortRef.current?.abort() }, [])

  return { poll, cancel }
}
