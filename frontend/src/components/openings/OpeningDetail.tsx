import styles from './OpeningDetail.module.css'
import { Badge } from '../ui/Badge'
import { Button } from '../ui/Button'
import { Icon } from '../ui/Icon'
import type { Opening } from '../../types/opening'
import { avatarColor, formatDate, formatSalary, initialsOf } from '../../utils/format'

interface OpeningDetailProps {
  opening: Opening
  onEdit: (opening: Opening) => void
  onDelete: (opening: Opening) => void
  onClose: () => void
}

export function OpeningDetail({ opening, onEdit, onDelete, onClose }: OpeningDetailProps) {
  return (
    <aside className={styles.panel} aria-label={`Detalhes da vaga ${opening.role}`}>
      <button className={styles.close} onClick={onClose} aria-label="Fechar detalhes">
        <Icon name="close" size={16} />
      </button>

      <header className={styles.header}>
        <div className={styles.avatar} style={{ backgroundColor: avatarColor(opening.company) }}>
          {initialsOf(opening.company)}
        </div>
        <h2 className={styles.role}>{opening.role}</h2>
        <p className={styles.company}>{opening.company}</p>

        <div className={styles.badges}>
          {opening.remote ? (
            <Badge tone="remote" icon={<Icon name="remote" size={13} />}>
              Remoto
            </Badge>
          ) : (
            <Badge tone="onsite" icon={<Icon name="building" size={13} />}>
              Presencial
            </Badge>
          )}
          <Badge tone="salary" icon={<Icon name="salary" size={13} />}>
            {formatSalary(opening.salary)}
          </Badge>
        </div>
      </header>

      <dl className={styles.facts}>
        <div className={styles.fact}>
          <dt>
            <Icon name="pin" size={15} />
            Localidade
          </dt>
          <dd>{opening.location}</dd>
        </div>
        <div className={styles.fact}>
          <dt>
            <Icon name="salary" size={15} />
            Remuneração
          </dt>
          <dd>{formatSalary(opening.salary)}</dd>
        </div>
        <div className={styles.fact}>
          <dt>
            <Icon name="clock" size={15} />
            Publicada em
          </dt>
          <dd>{formatDate(opening.createdAt)}</dd>
        </div>
        {opening.updatedAt !== opening.createdAt && (
          <div className={styles.fact}>
            <dt>
              <Icon name="edit" size={15} />
              Atualizada em
            </dt>
            <dd>{formatDate(opening.updatedAt)}</dd>
          </div>
        )}
      </dl>

      <div className={styles.actions}>
        <a
          className={styles.apply}
          href={opening.link}
          target="_blank"
          rel="noopener noreferrer"
        >
          <Icon name="external" size={16} />
          Ver vaga original
        </a>

        <div className={styles.secondary}>
          <Button size="sm" icon={<Icon name="edit" size={15} />} onClick={() => onEdit(opening)}>
            Editar
          </Button>
          <Button
            size="sm"
            variant="ghost"
            icon={<Icon name="trash" size={15} />}
            onClick={() => onDelete(opening)}
          >
            Excluir
          </Button>
        </div>
      </div>
    </aside>
  )
}
