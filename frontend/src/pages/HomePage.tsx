import { useMemo } from 'react'
import styles from './HomePage.module.css'

import { SearchForm } from '../components/openings/SearchForm'
import { StatsBar } from '../components/openings/StatsBar'
import { OpeningCard } from '../components/openings/OpeningCard'
import { SkeletonList } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'
import { Button } from '../components/ui/Button'
import { Icon } from '../components/ui/Icon'
import { ROUTE_PATHS } from '../router/useHashRoute'
import type { Filters } from '../components/openings/filters'
import type { Opening } from '../types/opening'

interface HomePageProps {
  openings: Opening[]
  loading: boolean
  filters: Filters
  locations: string[]
  onFiltersChange: (filters: Filters) => void
  onSearch: () => void
  onSelect: (opening: Opening) => void
  onEdit: (opening: Opening) => void
  onDelete: (opening: Opening) => void
  onCreate: () => void
}

const HIGHLIGHT_COUNT = 6

export function HomePage({
  openings,
  loading,
  filters,
  locations,
  onFiltersChange,
  onSearch,
  onSelect,
  onEdit,
  onDelete,
  onCreate,
}: HomePageProps) {
  const highlights = useMemo(
    () =>
      [...openings]
        .sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
        .slice(0, HIGHLIGHT_COUNT),
    [openings],
  )

  // The chips are drawn from the roles actually on the board, not a fixed list.
  const popular = useMemo(() => {
    const words = new Map<string, number>()
    for (const opening of openings) {
      for (const word of opening.role.split(/\s+/)) {
        if (word.length < 4) continue
        words.set(word, (words.get(word) ?? 0) + 1)
      }
    }
    return [...words.entries()]
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([word]) => word)
  }, [openings])

  return (
    <>
      <section className={styles.hero}>
        <div className={styles.heroInner}>
          <div className={styles.heroContent}>
            <p className={styles.eyebrow}>
              {openings.length > 0
                ? `${openings.length} ${openings.length === 1 ? 'vaga aberta' : 'vagas abertas'} agora`
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
              locations={locations}
              onChange={onFiltersChange}
              onSubmit={onSearch}
            />

            {popular.length > 0 && (
              <p className={styles.popular}>
                <span className={styles.popularLabel}>Buscas populares:</span>
                {popular.map((term) => (
                  <button
                    key={term}
                    className={styles.chip}
                    onClick={() => {
                      onFiltersChange({ ...filters, search: term })
                      onSearch()
                    }}
                  >
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
        {!loading && openings.length > 0 && (
          <section className={styles.section}>
            <StatsBar openings={openings} />
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
                <Button variant="primary" icon={<Icon name="plus" size={17} />} onClick={onCreate}>
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
                  onSelect={onSelect}
                  onEdit={onEdit}
                  onDelete={onDelete}
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

          <Button variant="primary" icon={<Icon name="send" size={17} />} onClick={onCreate}>
            Publicar vaga
          </Button>
        </section>
      </div>
    </>
  )
}
