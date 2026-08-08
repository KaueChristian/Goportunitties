import styles from './ResultsToolbar.module.css'
import { Icon } from '../ui/Icon'
import type { CardLayout } from './OpeningCard'
import type { SortOption } from './filters'

interface ResultsToolbarProps {
  count: number
  layout: CardLayout
  sort: SortOption
  onLayoutChange: (layout: CardLayout) => void
  onSortChange: (sort: SortOption) => void
}

const LAYOUTS: { value: CardLayout; label: string; icon: 'grid' | 'list' }[] = [
  { value: 'grid', label: 'Ver em grade', icon: 'grid' },
  { value: 'list', label: 'Ver em lista', icon: 'list' },
]

export function ResultsToolbar({
  count,
  layout,
  sort,
  onLayoutChange,
  onSortChange,
}: ResultsToolbarProps) {
  return (
    <div className={styles.toolbar}>
      <p className={styles.count} aria-live="polite">
        <strong>{count}</strong> {count === 1 ? 'vaga encontrada' : 'vagas encontradas'}
      </p>

      <div className={styles.controls}>
        <div className={styles.views} role="group" aria-label="Modo de exibição">
          {LAYOUTS.map((option) => (
            <button
              key={option.value}
              className={`${styles.view} ${layout === option.value ? styles.viewActive : ''}`}
              onClick={() => onLayoutChange(option.value)}
              aria-pressed={layout === option.value}
              aria-label={option.label}
              title={option.label}
            >
              <Icon name={option.icon} size={17} />
            </button>
          ))}
        </div>

        <label className={styles.sort}>
          <span className="sr-only">Ordenar vagas</span>
          <select value={sort} onChange={(event) => onSortChange(event.target.value as SortOption)}>
            <option value="recent">Mais recentes</option>
            <option value="salary-desc">Maior salário</option>
            <option value="salary-asc">Menor salário</option>
            <option value="role">Cargo (A–Z)</option>
          </select>
          <Icon name="chevron-down" size={16} />
        </label>
      </div>
    </div>
  )
}
