import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import { openingsApi } from '../api/openings'
import { ApiError } from '../api/client'
import { DEFAULT_FILTERS, sameResultSet, toQuery, type Filters } from '../components/openings/filters'
import { useDebouncedValue } from '../hooks/useDebouncedValue'
import { useOpeningsQuery } from '../hooks/useOpeningsQuery'
import type { Opening, OpeningFacets, OpeningPayload, OpeningStats, Pagination } from '../types/opening'

/** How many openings one page of results holds. */
export const PAGE_SIZE = 12

interface OpeningsContextValue {
  filters: Filters
  setFilters: (filters: Filters) => void
  patchFilters: (partial: Partial<Filters>) => void
  resetFilters: () => void

  page: number
  goToPage: (page: number) => void

  openings: Opening[]
  pagination: Pagination
  facets: OpeningFacets
  loading: boolean
  error: string | null
  reload: () => void

  stats: OpeningStats | null
  statsLoading: boolean

  selected: Opening | null
  select: (opening: Opening | null) => void

  /** Increments after every write, so dependent queries refetch. */
  revision: number

  createOpening: (payload: OpeningPayload) => Promise<Opening>
  updateOpening: (id: number, payload: OpeningPayload) => Promise<Opening>
  deleteOpening: (id: number) => Promise<void>
}

const OpeningsContext = createContext<OpeningsContextValue | null>(null)

/**
 * Owns everything about the openings collection: the filter state, the page on
 * screen, the aggregates and the writes.
 *
 * This used to live in App.tsx, which meant every page received the same ten
 * props whether it used them or not. A context is the smaller change than a
 * state library for an app this size, and it keeps the pages reading only what
 * they actually need.
 */
export function OpeningsProvider({ children }: { children: ReactNode }) {
  const [filters, setFiltersState] = useState<Filters>(DEFAULT_FILTERS)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<Opening | null>(null)
  const [revision, setRevision] = useState(0)

  const [stats, setStats] = useState<OpeningStats | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)

  // Typing filters the list, but not on every keystroke.
  const debouncedSearch = useDebouncedValue(filters.search)

  const query = useMemo(
    () => toQuery({ ...filters, search: debouncedSearch }, page, PAGE_SIZE),
    [filters, debouncedSearch, page],
  )

  const { openings, pagination, facets, loading, error, reload } = useOpeningsQuery(query, {
    facets: true,
    revision,
  })

  const setFilters = useCallback((next: Filters) => {
    setFiltersState((current) => {
      // Narrowing the set invalidates the current page; reordering does not.
      if (!sameResultSet(current, next)) setPage(1)
      return next
    })
  }, [])

  const patchFilters = useCallback(
    (partial: Partial<Filters>) => setFilters({ ...filters, ...partial }),
    [filters, setFilters],
  )

  const resetFilters = useCallback(() => {
    setFilters({ ...DEFAULT_FILTERS, sort: filters.sort })
  }, [filters.sort, setFilters])

  const goToPage = useCallback((next: number) => {
    setPage(next)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [])

  const invalidate = useCallback(() => setRevision((current) => current + 1), [])

  const createOpening = useCallback(
    async (payload: OpeningPayload) => {
      const created = await openingsApi.create(payload)
      invalidate()
      return created
    },
    [invalidate],
  )

  const updateOpening = useCallback(
    async (id: number, payload: OpeningPayload) => {
      const updated = await openingsApi.replace(id, payload)
      setSelected((current) => (current?.id === id ? updated : current))
      invalidate()
      return updated
    },
    [invalidate],
  )

  const deleteOpening = useCallback(
    async (id: number) => {
      await openingsApi.remove(id)
      setSelected((current) => (current?.id === id ? null : current))
      invalidate()
    },
    [invalidate],
  )

  // Aggregates cover the whole index, so they only change when a write happens.
  useEffect(() => {
    let active = true
    setStatsLoading(true)

    openingsApi
      .stats()
      .then((result) => {
        if (active) setStats(result)
      })
      .catch((err: unknown) => {
        if (!active) return
        // The dashboard degrades to "no numbers" rather than to an error page:
        // the listing below it is still perfectly usable.
        if (!(err instanceof ApiError)) throw err
        setStats(null)
      })
      .finally(() => {
        if (active) setStatsLoading(false)
      })

    return () => {
      active = false
    }
  }, [revision])

  const value = useMemo<OpeningsContextValue>(
    () => ({
      filters,
      setFilters,
      patchFilters,
      resetFilters,
      page,
      goToPage,
      openings,
      pagination,
      facets,
      loading,
      error,
      reload,
      stats,
      statsLoading,
      selected,
      select: setSelected,
      revision,
      createOpening,
      updateOpening,
      deleteOpening,
    }),
    [
      filters,
      setFilters,
      patchFilters,
      resetFilters,
      page,
      goToPage,
      openings,
      pagination,
      facets,
      loading,
      error,
      reload,
      stats,
      statsLoading,
      selected,
      revision,
      createOpening,
      updateOpening,
      deleteOpening,
    ],
  )

  return <OpeningsContext.Provider value={value}>{children}</OpeningsContext.Provider>
}

export function useOpenings(): OpeningsContextValue {
  const context = useContext(OpeningsContext)
  if (!context) {
    throw new Error('useOpenings precisa estar dentro de <OpeningsProvider>')
  }
  return context
}
