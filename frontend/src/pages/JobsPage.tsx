import { useMemo, useState } from 'react'
import styles from './JobsPage.module.css'

import { SearchForm } from '../components/openings/SearchForm'
import { FilterSidebar } from '../components/openings/FilterSidebar'
import { ResultsToolbar } from '../components/openings/ResultsToolbar'
import { OpeningCard, type CardLayout } from '../components/openings/OpeningCard'
import { OpeningDetail } from '../components/openings/OpeningDetail'
import { EmptyState } from '../components/ui/EmptyState'
import { SkeletonList } from '../components/ui/Skeleton'
import { Button } from '../components/ui/Button'
import { Icon } from '../components/ui/Icon'
import { countFacets, filterOpenings, isFiltered, DEFAULT_FILTERS } from '../components/openings/filters'
import type { Filters } from '../components/openings/filters'
import type { Opening } from '../types/opening'

interface JobsPageProps {
  openings: Opening[]
  loading: boolean
  error: string | null
  filters: Filters
  locations: string[]
  selected: Opening | null
  onFiltersChange: (filters: Filters) => void
  onSelect: (opening: Opening) => void
  onCloseDetail: () => void
  onEdit: (opening: Opening) => void
  onDelete: (opening: Opening) => void
  onCreate: () => void
  onReload: () => void
}

export function JobsPage({
  openings,
  loading,
  error,
  filters,
  locations,
  selected,
  onFiltersChange,
  onSelect,
  onCloseDetail,
  onEdit,
  onDelete,
  onCreate,
  onReload,
}: JobsPageProps) {
  const [layout, setLayout] = useState<CardLayout>('grid')

  const visible = useMemo(() => filterOpenings(openings, filters), [openings, filters])
  const counts = useMemo(() => countFacets(openings, filters), [openings, filters])

  return (
    <>
      <section className={styles.banner}>
        <div className={styles.bannerInner}>
          <h1 className={styles.title}>Vagas de tecnologia</h1>
          <p className={styles.lead}>
            Filtre por modalidade, salário e localidade para chegar às oportunidades que importam.
          </p>
          <SearchForm filters={filters} locations={locations} onChange={onFiltersChange} />
        </div>
      </section>

      <div className={`${styles.layout} ${selected ? styles.withDetail : ''}`}>
        {error ? (
          <div className={styles.errorArea}>
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
          </div>
        ) : (
          <>
            <FilterSidebar filters={filters} counts={counts} onChange={onFiltersChange} />

            <div className={styles.results}>
              <ResultsToolbar
                count={visible.length}
                layout={layout}
                sort={filters.sort}
                onLayoutChange={setLayout}
                onSortChange={(sort) => onFiltersChange({ ...filters, sort })}
              />

              {loading ? (
                <SkeletonList />
              ) : visible.length === 0 ? (
                isFiltered(filters) ? (
                  <EmptyState
                    icon="search"
                    title="Nenhuma vaga corresponde aos filtros"
                    description="Tente ajustar a busca, a modalidade, o salário mínimo ou a localidade."
                    action={
                      <Button
                        onClick={() => onFiltersChange({ ...DEFAULT_FILTERS, sort: filters.sort })}
                      >
                        Limpar filtros
                      </Button>
                    }
                  />
                ) : (
                  <EmptyState
                    title="Nenhuma vaga publicada ainda"
                    description="Publique a primeira oportunidade para começar a montar o radar de vagas."
                    action={
                      <Button
                        variant="primary"
                        icon={<Icon name="plus" size={17} />}
                        onClick={onCreate}
                      >
                        Publicar vaga
                      </Button>
                    }
                  />
                )
              ) : (
                <div className={`${styles.cards} ${styles[layout]}`}>
                  {visible.map((opening, index) => (
                    <OpeningCard
                      key={opening.id}
                      opening={opening}
                      index={index}
                      layout={layout}
                      selected={selected?.id === opening.id}
                      onSelect={onSelect}
                      onEdit={onEdit}
                      onDelete={onDelete}
                    />
                  ))}
                </div>
              )}
            </div>

            {selected && (
              <div className={styles.detail}>
                <OpeningDetail
                  opening={selected}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onClose={onCloseDetail}
                />
              </div>
            )}
          </>
        )}
      </div>
    </>
  )
}
