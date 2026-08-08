import type { FormEvent } from 'react'
import styles from './SearchForm.module.css'
import { Icon } from '../ui/Icon'
import type { Filters, RemoteFilter } from './filters'

interface SearchFormProps {
  filters: Filters
  locations: string[]
  /** Runs after the form is submitted — the hero uses it to jump to the results. */
  onSubmit?: () => void
  onChange: (filters: Filters) => void
}

export function SearchForm({ filters, locations, onSubmit, onChange }: SearchFormProps) {
  const patch = (partial: Partial<Filters>) => onChange({ ...filters, ...partial })

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    onSubmit?.()
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit} role="search">
      <div className={styles.field}>
        <Icon name="search" size={18} />
        <input
          type="search"
          value={filters.search}
          placeholder="Cargo ou empresa..."
          aria-label="Buscar por cargo ou empresa"
          onChange={(event) => patch({ search: event.target.value })}
        />
      </div>

      <div className={`${styles.field} ${styles.selectField}`}>
        <Icon name="pin" size={18} />
        <select
          value={filters.location}
          aria-label="Filtrar por localidade"
          onChange={(event) => patch({ location: event.target.value })}
        >
          <option value="all">Localidade</option>
          {locations.map((location) => (
            <option key={location} value={location}>
              {location}
            </option>
          ))}
        </select>
        <Icon name="chevron-down" size={16} />
      </div>

      <div className={`${styles.field} ${styles.selectField}`}>
        <Icon name="briefcase" size={18} />
        <select
          value={filters.remote}
          aria-label="Filtrar por modalidade"
          onChange={(event) => patch({ remote: event.target.value as RemoteFilter })}
        >
          <option value="all">Modalidade</option>
          <option value="remote">Remoto</option>
          <option value="onsite">Presencial</option>
        </select>
        <Icon name="chevron-down" size={16} />
      </div>

      <button className={styles.submit} type="submit">
        <Icon name="search" size={17} />
        Buscar
      </button>
    </form>
  )
}
