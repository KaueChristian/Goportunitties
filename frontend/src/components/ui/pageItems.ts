/** How many numbered buttons to show on each side of the current page. */
const WINDOW = 2

/** A page number, or the ellipsis standing in for a run of skipped pages. */
export type PageItem = number | 'gap'

/**
 * Builds the page buttons: the first and last page stay reachable, the pages
 * around the current one are listed, and everything else collapses into a gap.
 */
export function pageItems(page: number, totalPages: number): PageItem[] {
  if (totalPages < 1) return []

  const pages = new Set<number>([1, totalPages])
  for (let offset = -WINDOW; offset <= WINDOW; offset += 1) {
    const candidate = page + offset
    if (candidate >= 1 && candidate <= totalPages) pages.add(candidate)
  }

  const sorted = [...pages].sort((a, b) => a - b)
  const items: PageItem[] = []

  sorted.forEach((current, index) => {
    const previous = sorted[index - 1]

    if (index > 0 && current - previous === 2) {
      // Exactly one page is missing: showing it costs less room than the
      // ellipsis that would stand in for it, and it stays clickable.
      items.push(previous + 1)
    } else if (index > 0 && current - previous > 2) {
      items.push('gap')
    }

    items.push(current)
  })

  return items
}
