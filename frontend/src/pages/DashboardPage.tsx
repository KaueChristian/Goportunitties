import { useMemo, useState } from 'react'
import styles from './DashboardPage.module.css'

import { StatsBar } from '../components/openings/StatsBar'
import { OpeningCard } from '../components/openings/OpeningCard'
import { MonthlyChart } from '../components/dashboard/MonthlyChart'
import { EmptyState } from '../components/ui/EmptyState'
import { SkeletonList } from '../components/ui/Skeleton'
import { Button } from '../components/ui/Button'
import { Icon, type IconName } from '../components/ui/Icon'
import { ROUTE_PATHS } from '../router/useHashRoute'
import { formatRelativeDate, formatSalary } from '../utils/format'
import type { Opening } from '../types/opening'

type Panel = 'overview' | 'openings'

interface DashboardPageProps {
  openings: Opening[]
  loading: boolean
  error: string | null
  onSelect: (opening: Opening) => void
  onEdit: (opening: Opening) => void
  onDelete: (opening: Opening) => void
  onCreate: () => void
  onReload: () => void
}

const PANELS: { value: Panel; label: string; icon: IconName }[] = [
  { value: 'overview', label: 'Visão geral', icon: 'chart' },
  { value: 'openings', label: 'Vagas publicadas', icon: 'briefcase' },
]

const ACTIVITY_COUNT = 5
const RECENT_COUNT = 4

export function DashboardPage({
  openings,
  loading,
  error,
  onSelect,
  onEdit,
  onDelete,
  onCreate,
  onReload,
}: DashboardPageProps) {
  const [panel, setPanel] = useState<Panel>('overview')

  const byRecency = useMemo(
    () => [...openings].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt)),
    [openings],
  )

  // "Activity" is the real edit history the API already gives us: an opening
  // whose updatedAt moved past its createdAt was edited after publication.
  const activity = useMemo(
    () =>
      [...openings]
        .sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt))
        .slice(0, ACTIVITY_COUNT),
    [openings],
  )

  const topPaying = useMemo(
    () => [...openings].sort((a, b) => b.salary - a.salary).slice(0, 3),
    [openings],
  )

  return (
    <div className={styles.page}>
      <aside className={styles.sidebar}>
        <div className={styles.profile}>
          <span className={styles.avatar}>
            <Icon name="briefcase" size={28} />
          </span>
          {/* The page's h1: the panel has no title bar of its own to carry it. */}
          <h1 className={styles.profileName}>Goportunitties</h1>
          <p className={styles.profileRole}>Painel do índice de vagas</p>
        </div>

        <nav className={styles.nav} aria-label="Seções do painel">
          {PANELS.map((item) => (
            <button
              key={item.value}
              className={`${styles.navItem} ${panel === item.value ? styles.navItemActive : ''}`}
              onClick={() => setPanel(item.value)}
              aria-current={panel === item.value ? 'true' : undefined}
            >
              <Icon name={item.icon} size={18} />
              {item.label}
              {item.value === 'openings' && (
                <span className={styles.navCount}>{openings.length}</span>
              )}
            </button>
          ))}

          <button className={styles.navItem} onClick={onCreate}>
            <Icon name="send" size={18} />
            Publicar vaga
          </button>

          <a className={styles.navItem} href={ROUTE_PATHS.jobs}>
            <Icon name="search" size={18} />
            Buscar vagas
          </a>
        </nav>
      </aside>

      <div className={styles.content}>
        {error ? (
          <EmptyState
            icon="alert"
            tone="danger"
            title="Não foi possível carregar as vagas"
            description={error}
            action={
              <Button variant="primary" onClick={onReload}>
                Tentar novamente
              </Button>
            }
          />
        ) : loading ? (
          <SkeletonList count={3} />
        ) : panel === 'overview' ? (
          <>
            <StatsBar openings={openings} />

            <div className={styles.split}>
              <section className={styles.panel}>
                <header className={styles.panelHead}>
                  <h3 className={styles.panelTitle}>Publicações por mês</h3>
                  <span className={styles.panelNote}>Últimos 9 meses</span>
                </header>
                <MonthlyChart openings={openings} />
              </section>

              <section className={styles.panel}>
                <header className={styles.panelHead}>
                  <h3 className={styles.panelTitle}>Atividade recente</h3>
                </header>

                {activity.length === 0 ? (
                  <p className={styles.blank}>Nada por aqui ainda.</p>
                ) : (
                  <ul className={styles.activity}>
                    {activity.map((opening) => {
                      const edited = opening.updatedAt !== opening.createdAt
                      return (
                        <li key={opening.id}>
                          <span
                            className={`${styles.activityIcon} ${edited ? styles.activityEdited : ''}`}
                          >
                            <Icon name={edited ? 'edit' : 'plus'} size={15} />
                          </span>
                          <span className={styles.activityText}>
                            <strong>{opening.role}</strong> {edited ? 'atualizada' : 'publicada'} na{' '}
                            {opening.company}
                            <span className={styles.activityTime}>
                              {formatRelativeDate(edited ? opening.updatedAt : opening.createdAt)}
                            </span>
                          </span>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </section>
            </div>

            {topPaying.length > 0 && (
              <section className={styles.panel}>
                <header className={styles.panelHead}>
                  <h3 className={styles.panelTitle}>Maiores salários</h3>
                  <span className={styles.panelNote}>Do índice inteiro</span>
                </header>

                <ol className={styles.ranking}>
                  {topPaying.map((opening, index) => (
                    <li key={opening.id}>
                      <span className={styles.rankPosition}>{index + 1}</span>
                      <span className={styles.rankText}>
                        <strong>{opening.role}</strong>
                        <span>{opening.company}</span>
                      </span>
                      <span className={styles.rankSalary}>{formatSalary(opening.salary)}</span>
                    </li>
                  ))}
                </ol>
              </section>
            )}

            <section className={styles.panel}>
              <header className={styles.panelHead}>
                <h3 className={styles.panelTitle}>Publicadas recentemente</h3>
                <a className={styles.panelLink} href={ROUTE_PATHS.jobs}>
                  Ver todas
                  <Icon name="arrow-right" size={16} />
                </a>
              </header>

              {byRecency.length === 0 ? (
                <p className={styles.blank}>Nenhuma vaga publicada ainda.</p>
              ) : (
                <div className={styles.recent}>
                  {byRecency.slice(0, RECENT_COUNT).map((opening, index) => (
                    <OpeningCard
                      key={opening.id}
                      opening={opening}
                      index={index}
                      selected={false}
                      onSelect={onSelect}
                      onEdit={onEdit}
                      onDelete={onDelete}
                    />
                  ))}
                </div>
              )}
            </section>
          </>
        ) : byRecency.length === 0 ? (
          <EmptyState
            title="Nenhuma vaga publicada ainda"
            description="Publique a primeira oportunidade para começar a montar o radar de vagas."
            action={
              <Button variant="primary" icon={<Icon name="plus" size={17} />} onClick={onCreate}>
                Publicar vaga
              </Button>
            }
          />
        ) : (
          <div className={styles.stack}>
            {byRecency.map((opening, index) => (
              <OpeningCard
                key={opening.id}
                opening={opening}
                index={index}
                layout="list"
                selected={false}
                onSelect={onSelect}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
