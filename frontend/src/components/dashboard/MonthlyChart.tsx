import { useId, useMemo } from 'react'
import styles from './MonthlyChart.module.css'
import type { MonthCount } from '../../types/opening'
import { bucketByMonth } from './months'

interface MonthlyChartProps {
  /**
   * The series the API aggregated, as YYYY-MM. It only carries months that have
   * openings, so the chart still lays out the full axis itself and fills the
   * gaps with zeros.
   */
  data: MonthCount[]
  months?: number
}

const VIEW_WIDTH = 600
const VIEW_HEIGHT = 200
const GRID_LINES = 4

/**
 * Smooth line through the points using a cubic whose control points sit on the
 * horizontal midpoint between neighbours — the cheapest curve that reads as a
 * trend line without overshooting past the data.
 */
function buildPath(points: { x: number; y: number }[]): string {
  if (points.length === 0) return ''

  let path = `M ${points[0].x} ${points[0].y}`

  for (let i = 0; i < points.length - 1; i += 1) {
    const current = points[i]
    const next = points[i + 1]
    const midX = (current.x + next.x) / 2
    path += ` C ${midX} ${current.y}, ${midX} ${next.y}, ${next.x} ${next.y}`
  }

  return path
}

export function MonthlyChart({ data, months = 9 }: MonthlyChartProps) {
  const gradientId = useId()
  const buckets = useMemo(() => bucketByMonth(data, months), [data, months])

  const peak = Math.max(1, ...buckets.map((bucket) => bucket.count))
  const step = buckets.length > 1 ? VIEW_WIDTH / (buckets.length - 1) : 0

  const points = buckets.map((bucket, index) => ({
    x: index * step,
    y: VIEW_HEIGHT - (bucket.count / peak) * (VIEW_HEIGHT - 16),
  }))

  const line = buildPath(points)
  const area = `${line} L ${VIEW_WIDTH} ${VIEW_HEIGHT} L 0 ${VIEW_HEIGHT} Z`

  return (
    <figure className={styles.chart}>
      <figcaption className="sr-only">
        Vagas publicadas por mês nos últimos {months} meses. Pico de {peak} no período.
      </figcaption>

      <div className={styles.plot}>
        <svg
          className={styles.svg}
          viewBox={`0 0 ${VIEW_WIDTH} ${VIEW_HEIGHT}`}
          preserveAspectRatio="none"
          role="img"
          aria-label={buckets.map((bucket) => `${bucket.label}: ${bucket.count}`).join(', ')}
        >
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-accent)" stopOpacity="0.32" />
              <stop offset="100%" stopColor="var(--color-accent)" stopOpacity="0" />
            </linearGradient>
          </defs>

          {Array.from({ length: GRID_LINES }, (_, index) => {
            const y = (VIEW_HEIGHT / GRID_LINES) * index
            return (
              <line
                key={index}
                className={styles.grid}
                x1="0"
                y1={y}
                x2={VIEW_WIDTH}
                y2={y}
                vectorEffect="non-scaling-stroke"
              />
            )
          })}

          <path d={area} fill={`url(#${gradientId})`} />
          <path
            className={styles.line}
            d={line}
            fill="none"
            vectorEffect="non-scaling-stroke"
          />
        </svg>

        <span className={styles.peak} aria-hidden="true">
          {peak}
        </span>
      </div>

      <ul className={styles.axis} aria-hidden="true">
        {buckets.map((bucket, index) => (
          <li key={index}>{bucket.label}</li>
        ))}
      </ul>
    </figure>
  )
}
