import { describe, expect, it } from 'vitest'
import { pageItems } from './pageItems'

describe('pageItems', () => {
  it('lists every page while they still fit', () => {
    expect(pageItems(1, 5)).toEqual([1, 2, 3, 4, 5])
  })

  it('collapses the pages far from the current one', () => {
    expect(pageItems(10, 20)).toEqual([1, 'gap', 8, 9, 10, 11, 12, 'gap', 20])
  })

  it('keeps the first and last page reachable', () => {
    const items = pageItems(10, 20)

    expect(items[0]).toBe(1)
    expect(items.at(-1)).toBe(20)
  })

  it('does not open a gap of a single page', () => {
    // 1..3 and 4.. are adjacent, so an ellipsis hiding only page 4 would be
    // wider than the button it replaces.
    expect(pageItems(3, 8)).toEqual([1, 2, 3, 4, 5, 'gap', 8])
  })

  it('handles the edges', () => {
    expect(pageItems(1, 1)).toEqual([1])
    expect(pageItems(1, 0)).toEqual([])
  })
})
