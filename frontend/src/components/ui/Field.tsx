import { useId, type InputHTMLAttributes, type ReactNode } from 'react'
import styles from './Field.module.css'

// 'prefix' is a real (string) HTML attribute — omit it so we can reuse the name
// for a leading icon node.
interface FieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'id' | 'prefix'> {
  label: string
  error?: string
  hint?: string
  prefix?: ReactNode
}

export function Field({ label, error, hint, prefix, className, ...rest }: FieldProps) {
  const id = useId()
  const describedBy = error ? `${id}-error` : hint ? `${id}-hint` : undefined

  return (
    <div className={`${styles.field} ${className ?? ''}`}>
      <label className={styles.label} htmlFor={id}>
        {label}
      </label>

      <div className={`${styles.control} ${error ? styles.invalid : ''}`}>
        {prefix && <span className={styles.prefix}>{prefix}</span>}
        <input
          id={id}
          className={styles.input}
          aria-invalid={Boolean(error)}
          aria-describedby={describedBy}
          {...rest}
        />
      </div>

      {error ? (
        <p id={`${id}-error`} className={styles.error} role="alert">
          {error}
        </p>
      ) : (
        hint && (
          <p id={`${id}-hint`} className={styles.hint}>
            {hint}
          </p>
        )
      )}
    </div>
  )
}

interface ToggleProps {
  label: string
  description?: string
  checked: boolean
  onChange: (checked: boolean) => void
}

export function Toggle({ label, description, checked, onChange }: ToggleProps) {
  const id = useId()

  return (
    <div className={styles.toggleRow}>
      <div>
        <label className={styles.label} htmlFor={id}>
          {label}
        </label>
        {description && <p className={styles.hint}>{description}</p>}
      </div>

      <button
        id={id}
        type="button"
        role="switch"
        aria-checked={checked}
        aria-label={label}
        className={`${styles.toggle} ${checked ? styles.toggleOn : ''}`}
        onClick={() => onChange(!checked)}
      >
        <span className={styles.knob} />
      </button>
    </div>
  )
}
