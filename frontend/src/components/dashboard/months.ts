import type { MonthCount } from '../../types/opening'

const monthLabel = new Intl.DateTimeFormat('pt-BR', { month: 'short' })

export interface Bucket {
  label: string
  count: number
}

/**
 * Lays out a continuous axis of the last `months` months and drops the API's
 * series onto it.
 *
 * The aggregate only carries months that actually have openings, so the gaps
 * have to be filled here — otherwise a quiet month would silently disappear
 * from the axis and distort the trend line.
 */
export function bucketByMonth(data: MonthCount[], months: number): Bucket[] {
  const counts = new Map(data.map((entry) => [entry.month, entry.count]))
  const buckets: Bucket[] = []
  const now = new Date()

  for (let offset = months - 1; offset >= 0; offset -= 1) {
    const date = new Date(now.getFullYear(), now.getMonth() - offset, 1)
    // The API's key is YYYY-MM.
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`

    buckets.push({
      label: monthLabel.format(date).replace('.', ''),
      count: counts.get(key) ?? 0,
    })
  }

  return buckets
}
