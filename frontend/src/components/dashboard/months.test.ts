import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { bucketByMonth } from './months'

// Freeze the clock: "the last nine months" is otherwise a moving target.
beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date(2026, 7, 15)) // August 2026
})

afterEach(() => {
  vi.useRealTimers()
})

describe('bucketByMonth', () => {
  it('always lays out the full axis, however sparse the series is', () => {
    const buckets = bucketByMonth([{ month: '2026-08', count: 3 }], 9)

    expect(buckets).toHaveLength(9)
    expect(buckets.at(-1)?.count).toBe(3)
  })

  it('fills the months the API left out with zeros', () => {
    const buckets = bucketByMonth(
      [
        { month: '2026-06', count: 2 },
        { month: '2026-08', count: 5 },
      ],
      3,
    )

    expect(buckets.map((bucket) => bucket.count)).toEqual([2, 0, 5])
  })

  it('runs oldest to newest, the direction the chart is drawn', () => {
    const buckets = bucketByMonth(
      [
        { month: '2026-07', count: 1 },
        { month: '2026-08', count: 9 },
      ],
      2,
    )

    expect(buckets[0].count).toBe(1)
    expect(buckets[1].count).toBe(9)
  })

  it('ignores months outside the window', () => {
    const buckets = bucketByMonth([{ month: '2024-01', count: 99 }], 3)

    expect(buckets.every((bucket) => bucket.count === 0)).toBe(true)
  })

  it('matches the API key format, which zero-pads the month', () => {
    // A naive `${month + 1}` would produce "2026-9" and silently miss this one.
    vi.setSystemTime(new Date(2026, 8, 10)) // September 2026
    const buckets = bucketByMonth([{ month: '2026-09', count: 4 }], 1)

    expect(buckets[0].count).toBe(4)
  })

  it('is empty when no months are asked for', () => {
    expect(bucketByMonth([{ month: '2026-08', count: 3 }], 0)).toEqual([])
  })
})
