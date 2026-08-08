import styles from './Skeleton.module.css'

/** Placeholder block shown while the list loads, matching the card's rhythm. */
export function SkeletonCard() {
  return (
    <div className={styles.card} aria-hidden="true">
      <div className={styles.avatar} />
      <div className={styles.lines}>
        <div className={styles.line} style={{ width: '46%' }} />
        <div className={styles.line} style={{ width: '30%' }} />
        <div className={styles.chips}>
          <div className={styles.chip} />
          <div className={styles.chip} />
        </div>
      </div>
    </div>
  )
}

export function SkeletonList({ count = 4 }: { count?: number }) {
  return (
    <div className={styles.list} role="status" aria-label="Carregando vagas">
      {Array.from({ length: count }, (_, index) => (
        <SkeletonCard key={index} />
      ))}
    </div>
  )
}
