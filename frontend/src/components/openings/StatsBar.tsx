import styles from './StatsBar.module.css'
import { Icon, type IconName } from '../ui/Icon'
import type { Opening } from '../../types/opening'
import { formatSalaryCompact } from '../../utils/format'

type Tone = 'accent' | 'remote' | 'onsite'

interface StatsBarProps {
  openings: Opening[]
}

export function StatsBar({ openings }: StatsBarProps) {
  const total = openings.length
  const remote = openings.filter((opening) => opening.remote).length
  const companies = new Set(openings.map((opening) => opening.company)).size
  const averageSalary =
    total === 0
      ? 0
      : Math.round(openings.reduce((sum, opening) => sum + opening.salary, 0) / total)

  const tiles: { label: string; value: string; note?: string; icon: IconName; tone: Tone }[] = [
    {
      label: 'Vagas publicadas',
      value: String(total),
      note: companies > 0 ? `em ${companies} ${companies === 1 ? 'empresa' : 'empresas'}` : undefined,
      icon: 'briefcase',
      tone: 'accent',
    },
    {
      label: 'Oportunidades remotas',
      value: String(remote),
      note: total > 0 ? `${Math.round((remote / total) * 100)}% do total` : undefined,
      icon: 'remote',
      tone: 'remote',
    },
    {
      label: 'Salário médio',
      value: total > 0 ? formatSalaryCompact(averageSalary) : '—',
      note: total > 0 ? 'por mês' : undefined,
      icon: 'salary',
      tone: 'onsite',
    },
  ]

  return (
    <div className={styles.grid}>
      {tiles.map((tile) => (
        <div key={tile.label} className={`${styles.tile} ${styles[tile.tone]}`}>
          <div className={styles.text}>
            {/*
             * Large standalone figures use the font's proportional digits —
             * tabular-nums is for columns that must align, and looks loose here.
             */}
            <span className={styles.value}>{tile.value}</span>
            <span className={styles.label}>{tile.label}</span>
            {tile.note && <span className={styles.note}>{tile.note}</span>}
          </div>

          <span className={styles.icon}>
            <Icon name={tile.icon} size={24} />
          </span>
        </div>
      ))}
    </div>
  )
}
