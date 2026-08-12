import styles from './Pagination.module.css'
import { Icon } from './Icon'
import { pageItems } from './pageItems'

interface PaginationProps {
  page: number
  totalPages: number
  onChange: (page: number) => void
}

export function Pagination({ page, totalPages, onChange }: PaginationProps) {
  if (totalPages <= 1) return null

  const items = pageItems(page, totalPages)

  return (
    <nav className={styles.pagination} aria-label="Paginação dos resultados">
      <button
        className={styles.arrow}
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
        aria-label="Página anterior"
      >
        <Icon name="arrow-right" size={16} />
      </button>

      <ul className={styles.pages}>
        {items.map((item, index) =>
          item === 'gap' ? (
            <li key={`gap-${index}`} className={styles.gap} aria-hidden="true">
              …
            </li>
          ) : (
            <li key={item}>
              <button
                className={`${styles.page} ${item === page ? styles.current : ''}`}
                onClick={() => onChange(item)}
                aria-label={`Página ${item}`}
                aria-current={item === page ? 'page' : undefined}
              >
                {item}
              </button>
            </li>
          ),
        )}
      </ul>

      <button
        className={styles.arrow}
        onClick={() => onChange(page + 1)}
        disabled={page >= totalPages}
        aria-label="Próxima página"
      >
        <Icon name="arrow-right" size={16} />
      </button>
    </nav>
  )
}
