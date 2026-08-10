import { useCallback, useLayoutEffect, useState, type CSSProperties, type RefObject } from 'react'

/** Space between the anchor and the popup. */
const GAP = 8
/** Never squeeze the popup below this, even in a cramped viewport. */
const MIN_HEIGHT = 160
/** Breathing room against the edges of the window. */
const EDGE = 12

/**
 * Positions a popup against an anchor element, in viewport coordinates.
 *
 * A popup positioned inside its own subtree is at the mercy of every ancestor:
 * the hero band sets `overflow: hidden` to contain its decorative art, and that
 * clips any dropdown opened from the search bar. Rendering the popup in a
 * portal escapes the clipping, and `position: fixed` is what makes it land in
 * the right place once it no longer shares a containing block with its anchor.
 *
 * The height is capped by the room actually available, which is also what makes
 * the list's own scrollbar appear when the options do not fit.
 */
export function useAnchoredPosition(
  anchorRef: RefObject<HTMLElement | null>,
  open: boolean,
): CSSProperties {
  const [style, setStyle] = useState<CSSProperties>({ position: 'fixed' })

  const update = useCallback(() => {
    const anchor = anchorRef.current
    if (!anchor) return

    const rect = anchor.getBoundingClientRect()
    const below = window.innerHeight - rect.bottom - GAP - EDGE
    const above = rect.top - GAP - EDGE

    // Near the bottom of the window there is more room upwards; open there
    // rather than into a sliver of space.
    const flip = below < MIN_HEIGHT && above > below

    const width = rect.width
    const left = Math.max(EDGE, Math.min(rect.left, window.innerWidth - width - EDGE))

    setStyle({
      position: 'fixed',
      left,
      minWidth: width,
      maxHeight: Math.max(MIN_HEIGHT, flip ? above : below),
      ...(flip ? { bottom: window.innerHeight - rect.top + GAP } : { top: rect.bottom + GAP }),
    })
  }, [anchorRef])

  useLayoutEffect(() => {
    if (!open) return

    // Runs before paint, so the popup never shows up at the wrong coordinates.
    update()

    window.addEventListener('resize', update)
    // Capture phase: a scroll in any ancestor moves the anchor, not just the
    // one on the window.
    window.addEventListener('scroll', update, true)

    return () => {
      window.removeEventListener('resize', update)
      window.removeEventListener('scroll', update, true)
    }
  }, [open, update])

  return style
}
