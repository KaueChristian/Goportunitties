import type { ReactNode } from 'react'
import styles from './Badge.module.css'

type Tone = 'remote' | 'onsite' | 'salary' | 'neutral'

interface BadgeProps {
  tone?: Tone
  icon?: ReactNode
  children: ReactNode
}

export function Badge({ tone = 'neutral', icon, children }: BadgeProps) {
  return (
    <span className={`${styles.badge} ${styles[tone]}`}>
      {icon}
      {children}
    </span>
  )
}
