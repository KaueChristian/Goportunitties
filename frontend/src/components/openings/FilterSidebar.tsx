import styles from './FilterSidebar.module.css'
import { Icon } from '../ui/Icon'
import { formatSalary } from '../../utils/format'
import {
  DEFAULT_FILTERS,
  isFiltered,
  type FacetCounts,
  type Filters,
  type RemoteFilter,
} from './filters'

interface FilterSidebarProps {
  filters: Filters
  counts: FacetCounts
  onChange: (filters: Filters) => void
}

const REMOTE_OPTIONS: { value: RemoteFilter; label: string }[] = [
  { value: 'all', label: 'Todas' },
  { value: 'remote', label: 'Remoto' },
  { value: 'onsite', label: 'Presencial' },
]

const SALARY_STEP = 500

export function FilterSidebar({ filters, counts, onChange }: FilterSidebarProps) {
  const patch = (partial: Partial<Filters>) => onChange({ ...filters, ...partial })

  return (
    <aside className={styles.sidebar} aria-label="Filtros de vagas">
      <section className={styles.group}>
        <h3 className={styles.groupTitle}>Modalidade</h3>
        <ul className={styles.options}>
          {REMOTE_OPTIONS.map((option) => (
            <li key={option.value}>
              <label className={styles.option}>
                <input
                  className={styles.input}
                  type="radio"
                  name="modalidade"
                  checked={filters.remote === option.value}
                  onChange={() => patch({ remote: option.value })}
                />
                <span className={styles.box} aria-hidden="true">
                  <Icon name="check" size={12} />
                </span>
                <span className={styles.optionLabel}>{option.label}</span>
                <span className={styles.count}>{counts.remote[option.value]}</span>
              </label>
            </li>
          ))}
        </ul>
      </section>

      <section className={styles.group}>
        <h3 className={styles.groupTitle}>Salário mínimo</h3>
        <p className={styles.salary}>
          {filters.minSalary === 0 ? 'Qualquer valor' : `A partir de ${formatSalary(filters.minSalary)}`}
        </p>
        <input
          className={styles.range}
          type="range"
          min={0}
          max={counts.salaryCeiling}
          step={SALARY_STEP}
          value={Math.min(filters.minSalary, counts.salaryCeiling)}
          aria-label="Salário mínimo"
          onChange={(event) => patch({ minSalary: Number(event.target.value) })}
        />
        <div className={styles.rangeScale}>
          <span>{formatSalary(0)}</span>
          <span>{formatSalary(counts.salaryCeiling)}</span>
        </div>
      </section>

      <section className={styles.group}>
        <h3 className={styles.groupTitle}>Localidade</h3>
        <ul className={`${styles.options} ${styles.scrollable}`}>
          <li>
            <label className={styles.option}>
              <input
                className={styles.input}
                type="radio"
                name="localidade"
                checked={filters.location === 'all'}
                onChange={() => patch({ location: 'all' })}
              />
              <span className={styles.box} aria-hidden="true">
                <Icon name="check" size={12} />
              </span>
              <span className={styles.optionLabel}>Todas as localidades</span>
              <span className={styles.count}>
                {counts.locations.reduce((sum, item) => sum + item.count, 0)}
              </span>
            </label>
          </li>

          {counts.locations.map((location) => (
            <li key={location.value}>
              <label className={styles.option}>
                <input
                  className={styles.input}
                  type="radio"
                  name="localidade"
                  checked={filters.location === location.value}
                  onChange={() => patch({ location: location.value })}
                />
                <span className={styles.box} aria-hidden="true">
                  <Icon name="check" size={12} />
                </span>
                <span className={styles.optionLabel}>{location.value}</span>
                <span className={styles.count}>{location.count}</span>
              </label>
            </li>
          ))}
        </ul>
      </section>

      {isFiltered(filters) && (
        <button
          className={styles.clear}
          onClick={() => onChange({ ...DEFAULT_FILTERS, sort: filters.sort })}
        >
          <Icon name="close" size={15} />
          Limpar filtros
        </button>
      )}
    </aside>
  )
}
