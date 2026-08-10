import { useEffect, useId, useRef, useState, type KeyboardEvent, type ReactNode } from 'react'
import styles from './Select.module.css'
import { Icon } from './Icon'

export interface SelectOption {
  value: string
  label: string
  /** Optional trailing figure, e.g. how many openings the option would return. */
  count?: number
}

interface SelectProps {
  value: string
  options: SelectOption[]
  onChange: (value: string) => void
  /** Accessible name — the control has no visible <label>. */
  label: string
  icon?: ReactNode
  variant?: 'inline' | 'boxed'
  className?: string
}

/** How long a typed prefix keeps accumulating before it starts over. */
const TYPEAHEAD_RESET = 600

/**
 * A listbox built from real elements instead of a native <select>.
 *
 * The native popup is drawn by the browser, not the page: it ignores the design
 * tokens and, on Windows, renders a white sheet with a hard blue highlight over
 * a dark interface. Owning the markup is the only way to make the open state
 * look like the rest of the app.
 *
 * Focus stays on the trigger and the active option is announced through
 * aria-activedescendant, which is the pattern screen readers expect from a
 * collapsed combobox.
 */
export function Select({
  value,
  options,
  onChange,
  label,
  icon,
  variant = 'inline',
  className = '',
}: SelectProps) {
  const id = useId()
  const rootRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLUListElement>(null)
  const typed = useRef({ prefix: '', at: 0 })

  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(0)

  const selectedIndex = Math.max(
    0,
    options.findIndex((option) => option.value === value),
  )
  const selected = options[selectedIndex]

  // Clicking anywhere else closes the popup, which is what users expect from a
  // dropdown and what keeps it from covering the results underneath.
  useEffect(() => {
    if (!open) return

    const onPointerDown = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }

    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  // Keep the highlighted option inside the scroll area.
  useEffect(() => {
    if (!open) return
    listRef.current?.querySelector(`[data-index="${active}"]`)?.scrollIntoView({ block: 'nearest' })
  }, [open, active])

  const openAt = (index: number) => {
    setActive(index)
    setOpen(true)
  }

  const choose = (index: number) => {
    const option = options[index]
    if (option) onChange(option.value)
    setOpen(false)
  }

  const jumpToPrefix = (key: string) => {
    const now = Date.now()
    const prefix = now - typed.current.at > TYPEAHEAD_RESET ? key : typed.current.prefix + key
    typed.current = { prefix, at: now }

    const found = options.findIndex((option) => option.label.toLowerCase().startsWith(prefix))
    if (found >= 0) {
      setActive(found)
      if (!open) setOpen(true)
    }
  }

  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        if (open) setActive((current) => Math.min(current + 1, options.length - 1))
        else openAt(selectedIndex)
        return
      case 'ArrowUp':
        event.preventDefault()
        if (open) setActive((current) => Math.max(current - 1, 0))
        else openAt(selectedIndex)
        return
      case 'Home':
        if (!open) return
        event.preventDefault()
        setActive(0)
        return
      case 'End':
        if (!open) return
        event.preventDefault()
        setActive(options.length - 1)
        return
      case 'Enter':
      case ' ':
        event.preventDefault()
        if (open) choose(active)
        else openAt(selectedIndex)
        return
      case 'Escape':
        if (!open) return
        event.preventDefault()
        setOpen(false)
        return
      case 'Tab':
        setOpen(false)
        return
      default:
        // Any single printable character jumps to the matching option.
        if (event.key.length === 1 && !event.metaKey && !event.ctrlKey && !event.altKey) {
          jumpToPrefix(event.key.toLowerCase())
        }
    }
  }

  return (
    <div ref={rootRef} className={`${styles.root} ${className}`}>
      <button
        type="button"
        className={`${styles.trigger} ${styles[variant]}`}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={`${id}-list`}
        aria-label={label}
        aria-activedescendant={open ? `${id}-option-${active}` : undefined}
        onClick={() => {
          if (open) setOpen(false)
          else openAt(selectedIndex)
        }}
        onKeyDown={onKeyDown}
      >
        {icon}
        <span className={styles.value}>{selected?.label ?? label}</span>
        <Icon name="chevron-down" size={16} className={open ? styles.chevronOpen : styles.chevron} />
      </button>

      {open && (
        <ul ref={listRef} id={`${id}-list`} className={styles.list} role="listbox" aria-label={label}>
          {options.map((option, index) => (
            <li
              key={option.value}
              id={`${id}-option-${index}`}
              data-index={index}
              className={`${styles.option} ${index === active ? styles.optionActive : ''}`}
              role="option"
              aria-selected={option.value === value}
              // pointerdown, not click: the outside-click handler runs first
              // otherwise and the popup closes before the choice registers.
              onPointerDown={(event) => {
                event.preventDefault()
                choose(index)
              }}
              onPointerEnter={() => setActive(index)}
            >
              <span className={styles.optionLabel}>{option.label}</span>
              {option.count !== undefined && <span className={styles.count}>{option.count}</span>}
              {option.value === value && <Icon name="check" size={15} className={styles.check} />}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
