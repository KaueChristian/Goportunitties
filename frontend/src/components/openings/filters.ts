import type { OpeningsQuery } from '../../api/openings'

export type RemoteFilter = 'all' | 'remote' | 'onsite'
export type SortOption = 'recent' | 'salary-desc' | 'salary-asc' | 'role'

/**
 * The filter state the UI holds. Filtering itself happens on the server — this
 * type only describes the controls and how they translate into a query.
 */
export interface Filters {
  search: string
  location: string
  /** A source slug, or 'all'. */
  source: string
  remote: RemoteFilter
  sort: SortOption
  minSalary: number
}

export const DEFAULT_FILTERS: Filters = {
  search: '',
  location: 'all',
  source: 'all',
  remote: 'all',
  sort: 'recent',
  minSalary: 0,
}

/** Whether anything other than the sort order is narrowing the results. */
export function isFiltered(filters: Filters): boolean {
  return (
    filters.search !== '' ||
    filters.location !== 'all' ||
    filters.source !== 'all' ||
    filters.remote !== 'all' ||
    filters.minSalary > 0
  )
}

/** The modality radio has three states; the API takes an optional boolean. */
function toRemote(filter: RemoteFilter): boolean | undefined {
  if (filter === 'remote') return true
  if (filter === 'onsite') return false
  return undefined
}

/** Translates the UI's filter state into an API query. */
export function toQuery(filters: Filters, page = 1, pageSize?: number): OpeningsQuery {
  return {
    search: filters.search,
    location: filters.location,
    source: filters.source,
    remote: toRemote(filters.remote),
    minSalary: filters.minSalary,
    sort: filters.sort,
    page,
    pageSize,
  }
}

/**
 * Whether two filter sets would produce the same results. Used to decide when a
 * change should send the user back to page 1 — reordering keeps the page,
 * narrowing the set does not.
 */
export function sameResultSet(a: Filters, b: Filters): boolean {
  return (
    a.search === b.search &&
    a.location === b.location &&
    a.source === b.source &&
    a.remote === b.remote &&
    a.minSalary === b.minSalary
  )
}
