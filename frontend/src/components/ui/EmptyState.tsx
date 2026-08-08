import type { ReactNode } from 'react'
import styles from './EmptyState.module.css'
import { Icon, type IconName } from './Icon'

interface EmptyStateProps {
  icon?: IconName
  title: string
  description: string
  action?: ReactNode
  tone?: 'neutral' | 'danger'
}

export function EmptyState({
  icon = 'inbox',
  title,
  description,
  action,
  tone = 'neutral',
}: EmptyStateProps) {
  return (
    <div className={`${styles.wrapper} ${tone === 'danger' ? styles.danger : ''}`}>
      <div className={styles.iconRing}>
        <Icon name={icon} size={26} />
      </div>
      <h3 className={styles.title}>{title}</h3>
      <p className={styles.description}>{description}</p>
      {action && <div className={styles.action}>{action}</div>}
    </div>
  )
}
