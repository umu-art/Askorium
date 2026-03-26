import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { usePolling } from './usePolling'

/**
 * usePolling tests cover three main scenarios:
 *  1. Immediate resolution (isDone true on first call — no timer needed)
 *  2. Repeated polling until a condition is met
 *  3. Cancellation via AbortController
 *
 * Fake timers are used to avoid real waits and keep tests fast.
 * `vi.advanceTimersByTimeAsync` flushes both timers and pending microtasks,
 * which is necessary because the poll loop mixes Promise awaits and setTimeout.
 */
describe('usePolling', () => {
  beforeEach(() => { vi.useFakeTimers() })
  afterEach(() => { vi.useRealTimers() })

  it('resolves immediately when isDone returns true on first call', async () => {
    const { result } = renderHook(() => usePolling<number>(500))
    const fetcher = vi.fn().mockResolvedValue(42)
    const isDone = vi.fn().mockReturnValue(true)

    // No timer should be set — poll resolves after a single microtask
    const value = await result.current.poll(fetcher, isDone)

    expect(value).toBe(42)
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(isDone).toHaveBeenCalledWith(42)
  })

  it('keeps polling until isDone returns true', async () => {
    const { result } = renderHook(() => usePolling<string>(100))

    let calls = 0
    const fetcher = vi.fn().mockImplementation(() => Promise.resolve(`result-${++calls}`))
    // Returns true only on the third invocation
    const isDone = (v: string) => v === 'result-3'

    const pollPromise = result.current.poll(fetcher, isDone)

    // Two 100ms intervals separate the first two "not done" results.
    // advanceTimersByTimeAsync fires timers AND flushes microtasks after each.
    await vi.advanceTimersByTimeAsync(200)

    const final = await pollPromise
    expect(final).toBe('result-3')
    expect(fetcher).toHaveBeenCalledTimes(3)
  })

  it('rejects with AbortError when cancel() is called during timer wait', async () => {
    const { result } = renderHook(() => usePolling<number>(1000))
    const fetcher = vi.fn().mockResolvedValue(0)
    const isDone = vi.fn().mockReturnValue(false) // never finishes on its own

    const pollPromise = result.current.poll(fetcher, isDone)

    // Flush the first `await fetcher()` microtask so we enter the timer wait
    await vi.advanceTimersByTimeAsync(0)

    // Cancel while the 1000ms timer is still running
    result.current.cancel()

    await expect(pollPromise).rejects.toMatchObject({ name: 'AbortError' })
  })

  it('aborts the previous poll when poll() is called again', async () => {
    const { result } = renderHook(() => usePolling<string>(100))

    const neverDone = vi.fn().mockReturnValue(false)
    const alwaysDone = vi.fn().mockReturnValue(true)

    const p1 = result.current.poll(vi.fn().mockResolvedValue('a'), neverDone)
    const p2 = result.current.poll(vi.fn().mockResolvedValue('done'), alwaysDone)

    // p1 should be aborted by the start of p2
    await expect(p1).rejects.toMatchObject({ name: 'AbortError' })
    await expect(p2).resolves.toBe('done')
  })

  it('cancels in-flight polling on unmount', async () => {
    const { result, unmount } = renderHook(() => usePolling<number>(1000))
    const fetcher = vi.fn().mockResolvedValue(0)
    const isDone = vi.fn().mockReturnValue(false)

    const pollPromise = result.current.poll(fetcher, isDone)

    await vi.advanceTimersByTimeAsync(0) // past first fetch

    unmount() // triggers cleanup → abort

    await expect(pollPromise).rejects.toMatchObject({ name: 'AbortError' })
  })
})
