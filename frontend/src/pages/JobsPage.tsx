import { useState } from 'react'
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
import { Pagination } from '../components/ui/Pagination'
import { isFiltered } from '../components/openings/filters'
import { useOpenings } from '../state/OpeningsProvider'
import { useDialogs } from '../state/DialogsProvider'

export function JobsPage() {
  const {
    filters,
    setFilters,
    patchFilters,
    resetFilters,
    openings,
    pagination,
    facets,
    loading,
    error,
    reload,
    selected,
    select,
    page,
    goToPage,
  } = useOpenings()
  const { openCreate, openEdit, confirmDelete } = useDialogs()

  const [layout, setLayout] = useState<CardLayout>('grid')

  return (
    <>
      <section className={styles.banner}>
        <div className={styles.bannerInner}>
          <h1 className={styles.title}>Vagas de tecnologia</h1>
          <p className={styles.lead}>
            Filtre por modalidade, salário e localidade para chegar às oportunidades que importam.
          </p>
          <SearchForm
            filters={filters}
            locations={facets.locations.map((location) => location.value)}
            onChange={setFilters}
          />
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
                <Button variant="primary" onClick={reload}>
                  Tentar novamente
                </Button>
              }
            />
          </div>
        ) : (
          <>
            <FilterSidebar filters={filters} counts={facets} onChange={setFilters} />

            <div className={styles.results}>
              <ResultsToolbar
                count={pagination.total}
                layout={layout}
                sort={filters.sort}
                onLayoutChange={setLayout}
                onSortChange={(sort) => patchFilters({ sort })}
              />

              {loading ? (
                <SkeletonList />
              ) : openings.length === 0 ? (
                isFiltered(filters) ? (
                  <EmptyState
                    icon="search"
                    title="Nenhuma vaga corresponde aos filtros"
                    description="Tente ajustar a busca, a modalidade, o salário mínimo ou a localidade."
                    action={<Button onClick={resetFilters}>Limpar filtros</Button>}
                  />
                ) : (
                  <EmptyState
                    title="Nenhuma vaga publicada ainda"
                    description="Publique a primeira oportunidade para começar a montar o radar de vagas."
                    action={
                      <Button
                        variant="primary"
                        icon={<Icon name="plus" size={17} />}
                        onClick={openCreate}
                      >
                        Publicar vaga
                      </Button>
                    }
                  />
                )
              ) : (
                <>
                  <div className={`${styles.cards} ${styles[layout]}`}>
                    {openings.map((opening, index) => (
                      <OpeningCard
                        key={opening.id}
                        opening={opening}
                        index={index}
                        layout={layout}
                        selected={selected?.id === opening.id}
                        onSelect={select}
                        onEdit={openEdit}
                        onDelete={confirmDelete}
                      />
                    ))}
                  </div>

                  <Pagination
                    page={page}
                    totalPages={pagination.totalPages}
                    onChange={goToPage}
                  />
                </>
              )}
            </div>

            {selected && (
              <div className={styles.detail}>
                <OpeningDetail
                  opening={selected}
                  onEdit={openEdit}
                  onDelete={confirmDelete}
                  onClose={() => select(null)}
                />
              </div>
            )}
          </>
        )}
      </div>
    </>
  )
}
