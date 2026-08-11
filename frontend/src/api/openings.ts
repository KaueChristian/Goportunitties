import { api, toQueryString, type QueryParams } from './client'
import type {
  Opening,
  OpeningFacets,
  OpeningPayload,
  OpeningStats,
  Pagination,
  Suggestion,
} from '../types/opening'

/** Sort options the API accepts. `updated` powers the dashboard's activity feed. */
export type ListSort = 'recent' | 'salary-desc' | 'salary-asc' | 'role' | 'updated'

/** The filtering and paging surface of GET /openings. */
export interface OpeningsQuery {
  search?: string
  location?: string
  source?: string
  remote?: boolean
  minSalary?: number
  sort?: ListSort
  page?: number
  pageSize?: number
}

/** One page of results together with the window it represents. */
export interface OpeningsPage {
  openings: Opening[]
  pagination: Pagination
}

function toParams(query: OpeningsQuery): QueryParams {
  return {
    search: query.search?.trim() || undefined,
    // 'all' is how the UI spells "no filter" — it never goes on the wire.
    location: query.location && query.location !== 'all' ? query.location : undefined,
    source: query.source && query.source !== 'all' ? query.source : undefined,
    remote: query.remote,
    minSalary: query.minSalary && query.minSalary > 0 ? query.minSalary : undefined,
    sort: query.sort,
    page: query.page,
    pageSize: query.pageSize,
  }
}

/**
 * The only module that knows the shape of the openings endpoints. Components
 * and hooks talk to this, never to fetch directly.
 */
export const openingsApi = {
  async list(query: OpeningsQuery = {}): Promise<OpeningsPage> {
    const { data, meta } = await api.get<Opening[]>(`/openings${toQueryString(toParams(query))}`)
    const openings = data ?? []

    return {
      openings,
      // The API always sends meta; the fallback keeps a stubbed response usable.
      pagination: meta ?? {
        page: 1,
        pageSize: openings.length,
        total: openings.length,
        totalPages: 1,
      },
    }
  },

  async facets(query: OpeningsQuery = {}): Promise<OpeningFacets> {
    const { data } = await api.get<OpeningFacets>(
      `/openings/facets${toQueryString(toParams(query))}`,
    )
    return data
  },

  /** Roles and companies matching what the user typed, most frequent first. */
  async suggest(search: string, limit?: number): Promise<Suggestion[]> {
    const { data } = await api.get<Suggestion[]>(
      `/openings/suggestions${toQueryString({ search: search.trim(), limit })}`,
    )
    return data ?? []
  },

  async stats(): Promise<OpeningStats> {
    const { data } = await api.get<OpeningStats>('/openings/stats')
    return data
  },

  async show(id: number): Promise<Opening> {
    const { data } = await api.get<Opening>(`/openings/${id}`)
    return data
  },

  async create(payload: OpeningPayload): Promise<Opening> {
    const { data } = await api.post<Opening>('/openings', payload)
    return data
  },

  /** PUT: replaces the whole opening, so every field must be present. */
  async replace(id: number, payload: OpeningPayload): Promise<Opening> {
    const { data } = await api.put<Opening>(`/openings/${id}`, payload)
    return data
  },

  /** PATCH: sends only what changed — a zero or false included is applied. */
  async patch(id: number, partial: Partial<OpeningPayload>): Promise<Opening> {
    const { data } = await api.patch<Opening>(`/openings/${id}`, partial)
    return data
  },

  async remove(id: number): Promise<Opening> {
    const { data } = await api.delete<Opening>(`/openings/${id}`)
    return data
  },
}
