import styles from './StatsBar.module.css'
import { Icon, type IconName } from '../ui/Icon'
import type { OpeningStats } from '../../types/opening'
import { formatSalaryCompact } from '../../utils/format'

type Tone = 'accent' | 'remote' | 'onsite'

interface StatsBarProps {
  /**
   * Aggregates over the whole index, computed in SQL. They used to be derived
   * from the openings array on screen, which stopped being the whole set the
   * moment the listing was paginated.
   */
  stats: OpeningStats | null
}

export function StatsBar({ stats }: StatsBarProps) {
  const { total, remote, companies, averageSalary } = stats ?? {
    total: 0,
    remote: 0,
    companies: 0,
    averageSalary: 0,
  }

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
      value: averageSalary > 0 ? formatSalaryCompact(averageSalary) : '—',
      note: averageSalary > 0 ? 'por mês' : undefined,
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
