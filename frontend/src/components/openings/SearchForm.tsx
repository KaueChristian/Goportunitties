import { useEffect, useId, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import { createPortal } from 'react-dom'
import styles from './SearchForm.module.css'
import { Icon } from '../ui/Icon'
import { Select, type SelectOption } from '../ui/Select'
import { useAnchoredPosition } from '../../hooks/useAnchoredPosition'
import { useSuggestions } from '../../hooks/useSuggestions'
import type { Filters, RemoteFilter } from './filters'

interface SearchFormProps {
  filters: Filters
  locations: string[]
  /** Runs after the form is submitted — the hero uses it to jump to the results. */
  onSubmit?: () => void
  onChange: (filters: Filters) => void
}

const REMOTE_OPTIONS: SelectOption[] = [
  { value: 'all', label: 'Modalidade' },
  { value: 'remote', label: 'Remoto' },
  { value: 'onsite', label: 'Presencial' },
]

export function SearchForm({ filters, locations, onSubmit, onChange }: SearchFormProps) {
  const id = useId()
  const fieldRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLUListElement>(null)

  const [suggestionsOpen, setSuggestionsOpen] = useState(false)
  const [active, setActive] = useState(-1)

  const position = useAnchoredPosition(fieldRef, suggestionsOpen)

  const suggestions = useSuggestions(filters.search, suggestionsOpen)

  const patch = (partial: Partial<Filters>) => onChange({ ...filters, ...partial })

  const locationOptions: SelectOption[] = [
    { value: 'all', label: 'Localidade' },
    ...locations.map((location) => ({ value: location, label: location })),
  ]

  // A fresh list of suggestions has no highlighted row until the user picks one.
  useEffect(() => setActive(-1), [suggestions])

  useEffect(() => {
    if (!suggestionsOpen) return

    const onPointerDown = (event: PointerEvent) => {
      const target = event.target as Node
      // The list is portalled out of the field, so both count as "inside".
      if (fieldRef.current?.contains(target) || listRef.current?.contains(target)) return
      setSuggestionsOpen(false)
    }

    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [suggestionsOpen])

  const apply = (term: string) => {
    patch({ search: term })
    setSuggestionsOpen(false)
    onSubmit?.()
  }

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    setSuggestionsOpen(false)
    onSubmit?.()
  }

  const onSearchKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (!suggestionsOpen || suggestions.length === 0) return

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        setActive((current) => (current + 1) % suggestions.length)
        return
      case 'ArrowUp':
        event.preventDefault()
        setActive((current) => (current <= 0 ? suggestions.length - 1 : current - 1))
        return
      case 'Enter':
        // Only intercept when a suggestion is highlighted; otherwise the form
        // submits with whatever was typed, which is the expected default.
        if (active >= 0) {
          event.preventDefault()
          apply(suggestions[active].value)
        }
        return
      case 'Escape':
        event.preventDefault()
        setSuggestionsOpen(false)
    }
  }

  const listId = `${id}-suggestions`

  return (
    <form className={styles.form} onSubmit={handleSubmit} role="search">
      <div ref={fieldRef} className={`${styles.field} ${styles.searchField}`}>
        <Icon name="search" size={18} />
        <input
          type="search"
          value={filters.search}
          placeholder="Cargo ou empresa..."
          aria-label="Buscar por cargo ou empresa"
          role="combobox"
          aria-expanded={suggestionsOpen && suggestions.length > 0}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={active >= 0 ? `${listId}-${active}` : undefined}
          autoComplete="off"
          onChange={(event) => {
            patch({ search: event.target.value })
            setSuggestionsOpen(true)
          }}
          onFocus={() => setSuggestionsOpen(true)}
          onKeyDown={onSearchKeyDown}
        />

        {suggestionsOpen &&
          suggestions.length > 0 &&
          createPortal(
            <ul
              ref={listRef}
              id={listId}
              className={styles.suggestions}
              style={position}
              role="listbox"
              aria-label="Sugestões de busca"
            >
              {suggestions.map((suggestion, index) => (
                <li
                  key={`${suggestion.kind}-${suggestion.value}`}
                  id={`${listId}-${index}`}
                  className={`${styles.suggestion} ${index === active ? styles.suggestionActive : ''}`}
                  role="option"
                  aria-selected={index === active}
                  // pointerdown, not click: the input's blur would otherwise
                  // close the list before the choice registers.
                  onPointerDown={(event) => {
                    event.preventDefault()
                    apply(suggestion.value)
                  }}
                  onPointerEnter={() => setActive(index)}
                >
                  <Icon name={suggestion.kind === 'company' ? 'building' : 'briefcase'} size={15} />
                  <span className={styles.suggestionValue}>{suggestion.value}</span>
                  <span className={styles.suggestionCount}>
                    {suggestion.count} {suggestion.count === 1 ? 'vaga' : 'vagas'}
                  </span>
                </li>
              ))}
            </ul>,
            document.body,
          )}
      </div>

      <Select
        className={styles.field}
        value={filters.location}
        options={locationOptions}
        onChange={(location) => patch({ location })}
        label="Filtrar por localidade"
        icon={<Icon name="pin" size={18} />}
      />

      <Select
        className={styles.field}
        value={filters.remote}
        options={REMOTE_OPTIONS}
        onChange={(remote) => patch({ remote: remote as RemoteFilter })}
        label="Filtrar por modalidade"
        icon={<Icon name="briefcase" size={18} />}
      />

      <button className={styles.submit} type="submit">
        <Icon name="search" size={17} />
        Buscar
      </button>
    </form>
  )
}
