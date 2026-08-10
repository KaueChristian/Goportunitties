import { useMemo } from 'react'
import styles from './HomePage.module.css'

import { SearchForm } from '../components/openings/SearchForm'
import { StatsBar } from '../components/openings/StatsBar'
import { OpeningCard } from '../components/openings/OpeningCard'
import { SkeletonList } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'
import { Button } from '../components/ui/Button'
import { Icon } from '../components/ui/Icon'
import { ROUTE_PATHS, useHashRoute } from '../router/useHashRoute'
import { useOpeningsQuery } from '../hooks/useOpeningsQuery'
import { useOpenings } from '../state/OpeningsProvider'
import { useDialogs } from '../state/DialogsProvider'
import type { Opening } from '../types/opening'

const HIGHLIGHT_COUNT = 6
/** Fetched beyond what is shown, so the popular-term chips have something to count. */
const HIGHLIGHT_FETCH = 24

export function HomePage() {
  const { navigate } = useHashRoute()
  const { filters, setFilters, facets, stats, statsLoading, select, revision } = useOpenings()
  const { openCreate, openEdit, confirmDelete } = useDialogs()

  // The home page always shows the newest openings, whatever the jobs page is
  // currently filtered by — so it runs its own query.
  const { openings: recent, loading } = useOpeningsQuery(
    { sort: 'recent', page: 1, pageSize: HIGHLIGHT_FETCH },
    { revision },
  )

  const highlights = recent.slice(0, HIGHLIGHT_COUNT)
  const total = stats?.total ?? 0

  // The chips are drawn from the roles actually on the board, not a fixed list.
  const popular = useMemo(() => {
    const words = new Map<string, number>()
    for (const opening of recent) {
      for (const word of opening.role.split(/\s+/)) {
        if (word.length < 4) continue
        words.set(word, (words.get(word) ?? 0) + 1)
      }
    }
    return [...words.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([word]) => word)
  }, [recent])

  const openDetail = (opening: Opening) => {
    select(opening)
    navigate('jobs')
  }

  const searchFor = (term: string) => {
    setFilters({ ...filters, search: term })
    navigate('jobs')
  }

  return (
    <>
      <section className={styles.hero}>
        <div className={styles.heroInner}>
          <div className={styles.heroContent}>
            <p className={styles.eyebrow}>
              {total > 0
                ? `${total} ${total === 1 ? 'vaga aberta' : 'vagas abertas'} agora`
                : 'Seu radar de oportunidades'}
            </p>

            <h1 className={styles.title}>
              Encontre a vaga que <mark className={styles.mark}>combina</mark> com a sua carreira
            </h1>

            <p className={styles.lead}>
              Um índice aberto de oportunidades em tecnologia. Filtre por modalidade, salário e
              localidade — e publique as suas próprias vagas em segundos.
            </p>

            <SearchForm
              filters={filters}
              locations={facets.locations.map((location) => location.value)}
              onChange={setFilters}
              onSubmit={() => navigate('jobs')}
            />

            {popular.length > 0 && (
              <p className={styles.popular}>
                <span className={styles.popularLabel}>Buscas populares:</span>
                {popular.map((term) => (
                  <button key={term} className={styles.chip} onClick={() => searchFor(term)}>
                    {term}
                  </button>
                ))}
              </p>
            )}
          </div>

          {/*
           * A composed stack of cards instead of a stock photo: it says "job
           * board" with the same shapes the rest of the page already uses, and
           * ships nothing extra over the wire.
           */}
          <div className={styles.heroArt} aria-hidden="true">
            <span className={styles.blob} />
            <div className={`${styles.artCard} ${styles.artCard1}`}>
              <span className={styles.artLogo} />
              <span className={styles.artLines}>
                <i />
                <i />
              </span>
            </div>
            <div className={`${styles.artCard} ${styles.artCard2}`}>
              <span className={`${styles.artLogo} ${styles.artLogoAlt}`} />
              <span className={styles.artLines}>
                <i />
                <i />
              </span>
            </div>
            <div className={`${styles.artCard} ${styles.artCard3}`}>
              <span className={`${styles.artLogo} ${styles.artLogoAlt2}`} />
              <span className={styles.artLines}>
                <i />
                <i />
              </span>
            </div>
            <span className={styles.dots} />
          </div>
        </div>
      </section>

      <div className={styles.page}>
        {!statsLoading && total > 0 && (
          <section className={styles.section}>
            <StatsBar stats={stats} />
          </section>
        )}

        <section className={styles.section}>
          <header className={styles.sectionHead}>
            <div>
              <p className={styles.sectionTagline}>Publicadas recentemente</p>
              <h2 className={styles.sectionTitle}>Vagas em destaque</h2>
            </div>

            <a className={styles.sectionLink} href={ROUTE_PATHS.jobs}>
              Ver todas as vagas
              <Icon name="arrow-right" size={17} />
            </a>
          </header>

          {loading ? (
            <SkeletonList count={3} />
          ) : highlights.length === 0 ? (
            <EmptyState
              title="Nenhuma vaga publicada ainda"
              description="Publique a primeira oportunidade para começar a montar o radar de vagas."
              action={
                <Button variant="primary" icon={<Icon name="plus" size={17} />} onClick={openCreate}>
                  Publicar vaga
                </Button>
              }
            />
          ) : (
            <div className={styles.grid}>
              {highlights.map((opening, index) => (
                <OpeningCard
                  key={opening.id}
                  opening={opening}
                  index={index}
                  selected={false}
                  onSelect={openDetail}
                  onEdit={openEdit}
                  onDelete={confirmDelete}
                />
              ))}
            </div>
          )}
        </section>

        <section className={styles.cta}>
          <div className={styles.ctaText}>
            <h2 className={styles.ctaTitle}>Está contratando?</h2>
            <p className={styles.ctaLead}>
              Publique a vaga no índice e ela aparece na busca imediatamente.
            </p>
          </div>

          <Button variant="primary" icon={<Icon name="send" size={17} />} onClick={openCreate}>
            Publicar vaga
          </Button>
        </section>
      </div>
    </>
  )
}
