import type { Opening } from '../../types/opening'
import { normalize } from '../../utils/format'

export type RemoteFilter = 'all' | 'remote' | 'onsite'
export type SortOption = 'recent' | 'salary-desc' | 'salary-asc' | 'role'

export interface Filters {
  search: string
  location: string
  remote: RemoteFilter
  sort: SortOption
  minSalary: number
}

export const DEFAULT_FILTERS: Filters = {
  search: '',
  location: 'all',
  remote: 'all',
  sort: 'recent',
  minSalary: 0,
}

/** Which criterion to ignore — used to count a facet against everything but itself. */
type Facet = 'search' | 'location' | 'remote' | 'salary'

function matches(opening: Opening, filters: Filters, except?: Facet): boolean {
  if (except !== 'remote') {
    if (filters.remote === 'remote' && !opening.remote) return false
    if (filters.remote === 'onsite' && opening.remote) return false
  }

  if (except !== 'location' && filters.location !== 'all' && opening.location !== filters.location) {
    return false
  }

  if (except !== 'salary' && opening.salary < filters.minSalary) return false

  if (except !== 'search') {
    const term = normalize(filters.search.trim())
    if (term && !normalize(`${opening.role} ${opening.company}`).includes(term)) return false
  }

  return true
}

export function filterOpenings(openings: Opening[], filters: Filters): Opening[] {
  const visible = openings.filter((opening) => matches(opening, filters))

  switch (filters.sort) {
    case 'salary-desc':
      return visible.sort((a, b) => b.salary - a.salary)
    case 'salary-asc':
      return visible.sort((a, b) => a.salary - b.salary)
    case 'role':
      return visible.sort((a, b) => a.role.localeCompare(b.role))
    default:
      return visible.sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
  }
}

export function isFiltered(filters: Filters): boolean {
  return (
    filters.search !== '' ||
    filters.location !== 'all' ||
    filters.remote !== 'all' ||
    filters.minSalary > 0
  )
}

export interface FacetCounts {
  remote: Record<RemoteFilter, number>
  locations: { value: string; count: number }[]
  salaryCeiling: number
}

/**
 * Counts shown beside each option. Every facet is counted against the list
 * filtered by the *other* criteria, so picking one option never leaves its
 * siblings showing a total the click can't produce.
 */
export function countFacets(openings: Opening[], filters: Filters): FacetCounts {
  const forRemote = openings.filter((opening) => matches(opening, filters, 'remote'))
  const forLocation = openings.filter((opening) => matches(opening, filters, 'location'))

  const byLocation = new Map<string, number>()
  for (const opening of forLocation) {
    byLocation.set(opening.location, (byLocation.get(opening.location) ?? 0) + 1)
  }

  return {
    remote: {
      all: forRemote.length,
      remote: forRemote.filter((opening) => opening.remote).length,
      onsite: forRemote.filter((opening) => !opening.remote).length,
    },
    locations: [...byLocation.entries()]
      .map(([value, count]) => ({ value, count }))
      .sort((a, b) => a.value.localeCompare(b.value)),
    // Rounded up to a clean step so the slider's max isn't an odd salary figure.
    salaryCeiling: Math.max(1000, Math.ceil(Math.max(0, ...openings.map((o) => o.salary)) / 1000) * 1000),
  }
}
