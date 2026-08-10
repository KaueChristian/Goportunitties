import { useCallback, useEffect, useState } from 'react'

export type Theme = 'light' | 'dark'

const STORAGE_KEY = 'goportunitties:theme'

/**
 * How long the cross-fade between themes lasts. Must stay in step with the
 * duration in global.css — the attribute is removed once the paint is done, so
 * a value shorter than the CSS one would cut the animation off mid-way.
 */
const TRANSITION_MS = 320
const TRANSITION_ATTR = 'data-theme-transition'

function initialTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

/** Survives re-renders so a fast double toggle does not end the fade early. */
let transitionTimer = 0

/** Stamps data-theme on <html>, which the token layer keys off of. */
export function useTheme() {
  const [theme, setTheme] = useState<Theme>(initialTheme)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    localStorage.setItem(STORAGE_KEY, theme)
  }, [theme])

  const toggleTheme = useCallback(() => {
    // The cross-fade is switched on only around the toggle. Leaving those
    // transitions permanently on would make every hover and focus state fade
    // too, which reads as lag on controls that should feel instant.
    const root = document.documentElement
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    if (!reduced) {
      window.clearTimeout(transitionTimer)
      root.setAttribute(TRANSITION_ATTR, '')
      transitionTimer = window.setTimeout(() => root.removeAttribute(TRANSITION_ATTR), TRANSITION_MS)
    }

    setTheme((current) => (current === 'dark' ? 'light' : 'dark'))
  }, [])

  return { theme, toggleTheme }
}
