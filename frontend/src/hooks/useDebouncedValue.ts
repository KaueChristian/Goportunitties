import { useEffect, useState } from 'react'

/**
 * Trails `value` by `delay` milliseconds, resetting the timer on every change.
 *
 * The search box is the reason this exists: without it, every keystroke would
 * be one request to the API.
 */
export function useDebouncedValue<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value), delay)
    return () => window.clearTimeout(timer)
  }, [value, delay])

  return debounced
}
