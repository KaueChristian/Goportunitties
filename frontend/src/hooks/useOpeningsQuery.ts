import { useCallback, useEffect, useRef, useState } from 'react'
import { openingsApi, type OpeningsQuery } from '../api/openings'
import { ApiError } from '../api/client'
import type { Opening, OpeningFacets, Pagination } from '../types/opening'

const EMPTY_PAGINATION: Pagination = { page: 1, pageSize: 0, total: 0, totalPages: 0 }

const EMPTY_FACETS: OpeningFacets = {
  remote: { all: 0, remote: 0, onsite: 0 },
  locations: [],
  sources: [],
  salaryCeiling: 1000,
}

interface Options {
  /** Also fetch the facet counts. Only the filter sidebar needs them. */
  facets?: boolean
  /** Bump to force a refetch — the provider uses it to invalidate after a write. */
  revision?: number
}

export interface OpeningsQueryResult {
  openings: Opening[]
  pagination: Pagination
  facets: OpeningFacets
  loading: boolean
  error: string | null
  reload: () => void
}

/**
 * Fetches one page of openings for a query, plus its facets when asked.
 *
 * Now that filtering, sorting and paging happen in SQL, the query object is the
 * single input: change it and the hook refetches. Results are keyed by the
 * serialized query so a slow response for an abandoned filter cannot overwrite
 * the current one.
 */
export function useOpeningsQuery(query: OpeningsQuery, options: Options = {}): OpeningsQueryResult {
  const { facets: wantFacets = false, revision = 0 } = options

  const [openings, setOpenings] = useState<Opening[]>([])
  const [pagination, setPagination] = useState<Pagination>(EMPTY_PAGINATION)
  const [facets, setFacets] = useState<OpeningFacets>(EMPTY_FACETS)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // The query is an object literal at every render, so compare it by value.
  const key = JSON.stringify(query)
  const latestKey = useRef(key)

  const load = useCallback(async () => {
    latestKey.current = key
    setLoading(true)
    setError(null)

    const parsed = JSON.parse(key) as OpeningsQuery

    try {
      const [page, counts] = await Promise.all([
        openingsApi.list(parsed),
        wantFacets ? openingsApi.facets(parsed) : Promise.resolve(null),
      ])

      // A response for a filter the user has already moved on from is discarded.
      if (latestKey.current !== key) return

      setOpenings(page.openings)
      setPagination(page.pagination)
      if (counts) setFacets(counts)
      setError(null)
    } catch (err) {
      if (latestKey.current !== key) return

      setOpenings([])
      setPagination(EMPTY_PAGINATION)
      setError(
        err instanceof ApiError ? err.message : 'Erro inesperado ao carregar as vagas.',
      )
    } finally {
      if (latestKey.current === key) setLoading(false)
    }
  }, [key, wantFacets])

  useEffect(() => {
    void load()
  }, [load, revision])

  return { openings, pagination, facets, loading, error, reload: () => void load() }
}
