import styles from './OpeningCard.module.css'
import { Icon } from '../ui/Icon'
import type { Opening } from '../../types/opening'
import { avatarColor, formatRelativeDate, formatSalary, initialsOf } from '../../utils/format'

export type CardLayout = 'grid' | 'list'

interface OpeningCardProps {
  opening: Opening
  selected: boolean
  index: number
  layout?: CardLayout
  onSelect: (opening: Opening) => void
  onEdit: (opening: Opening) => void
  onDelete: (opening: Opening) => void
}

export function OpeningCard({
  opening,
  selected,
  index,
  layout = 'grid',
  onSelect,
  onEdit,
  onDelete,
}: OpeningCardProps) {
  return (
    <article
      className={`${styles.card} ${styles[layout]} ${selected ? styles.selected : ''}`}
      // Staggered entrance, capped so a long list doesn't wait on the last item.
      style={{ animationDelay: `${Math.min(index, 8) * 40}ms` }}
    >
      {/*
       * The whole card selects the opening via a stretched transparent button, so
       * the accessible name lives on a real control while the apply link and the
       * row actions stay above it.
       */}
      <button
        className={styles.hitArea}
        onClick={() => onSelect(opening)}
        aria-label={`Ver detalhes da vaga ${opening.role} na ${opening.company}`}
        aria-current={selected}
      />

      <div className={styles.top}>
        <div className={styles.logo} style={{ backgroundColor: avatarColor(opening.company) }}>
          {initialsOf(opening.company)}
        </div>

        <div className={styles.headings}>
          <h3 className={styles.role}>{opening.role}</h3>
          <p className={styles.company}>
            <Icon name="building" size={15} />
            {opening.company}
          </p>
        </div>

        <div className={styles.actions}>
          <button
            className={styles.action}
            onClick={() => onEdit(opening)}
            aria-label={`Editar vaga ${opening.role}`}
            title="Editar"
          >
            <Icon name="edit" size={16} />
          </button>
          <button
            className={`${styles.action} ${styles.destructive}`}
            onClick={() => onDelete(opening)}
            aria-label={`Excluir vaga ${opening.role}`}
            title="Excluir"
          >
            <Icon name="trash" size={16} />
          </button>
        </div>
      </div>

      <ul className={styles.info}>
        <li className={opening.remote ? styles.remote : styles.onsite}>
          <Icon name={opening.remote ? 'remote' : 'building'} size={15} />
          {opening.remote ? 'Remoto' : 'Presencial'}
        </li>
        <li>
          <Icon name="clock" size={15} />
          {formatRelativeDate(opening.createdAt)}
        </li>
        <li className={styles.location}>
          <Icon name="pin" size={15} />
          {opening.location}
        </li>
      </ul>

      <div className={styles.bottom}>
        <p className={styles.salary}>
          {formatSalary(opening.salary)}
          <span className={styles.salaryUnit}>/mês</span>
        </p>

        <a
          className={styles.apply}
          href={opening.link}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(event) => event.stopPropagation()}
        >
          Ver vaga
          <Icon name="external" size={15} />
        </a>
      </div>
    </article>
  )
}
