import { useEffect, useRef, useState } from 'react'
import { openingsApi } from '../api/openings'
import { useDebouncedValue } from './useDebouncedValue'
import type { Suggestion } from '../types/opening'

/** Below this, almost every opening matches — the API answers nothing anyway. */
export const MIN_TERM_LENGTH = 2

/**
 * Roles and companies matching `term`, for the search box's autocomplete.
 *
 * The term is debounced so a word typed at speed costs one request instead of
 * one per key, and responses are keyed by the term that asked for them: a slow
 * answer for an abandoned prefix must never replace a newer list.
 */
export function useSuggestions(term: string, enabled = true): Suggestion[] {
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])
  const debounced = useDebouncedValue(term, 200)
  const latest = useRef('')

  useEffect(() => {
    const trimmed = debounced.trim()
    latest.current = trimmed

    if (!enabled || trimmed.length < MIN_TERM_LENGTH) {
      setSuggestions([])
      return
    }

    let active = true

    openingsApi
      .suggest(trimmed)
      .then((found) => {
        if (active && latest.current === trimmed) setSuggestions(found)
      })
      .catch(() => {
        // The search box still works without suggestions; a failure here is not
        // worth an error message.
        if (active) setSuggestions([])
      })

    return () => {
      active = false
    }
  }, [debounced, enabled])

  return suggestions
}
