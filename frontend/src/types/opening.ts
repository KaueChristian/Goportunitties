/** Mirrors dto.OpeningResponse on the Go side. */
export interface Opening {
  id: number
  createdAt: string
  updatedAt: string
  role: string
  company: string
  location: string
  remote: boolean
  link: string
  salary: number
}

/** Mirrors dto.OpeningRequest — the full representation POST and PUT expect. */
export interface OpeningPayload {
  role: string
  company: string
  location: string
  remote: boolean
  link: string
  salary: number
}

/** Mirrors dto.Pagination — the window `data` represents. */
export interface Pagination {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

/** Mirrors dto.RemoteCounts. */
export interface RemoteCounts {
  all: number
  remote: number
  onsite: number
}

/** Mirrors dto.LocationCount. */
export interface LocationCount {
  value: string
  count: number
}

/**
 * Mirrors dto.OpeningFacets. Each count is taken against the current query with
 * that facet removed, which is why the numbers are computed server-side rather
 * than from the page of results on screen.
 */
export interface OpeningFacets {
  remote: RemoteCounts
  locations: LocationCount[]
  salaryCeiling: number
}

/** Mirrors dto.MonthCount — one bar of the dashboard chart, as YYYY-MM. */
export interface MonthCount {
  month: string
  count: number
}

/** Mirrors dto.OpeningStats — aggregates over the whole index, not one page. */
export interface OpeningStats {
  total: number
  remote: number
  onsite: number
  companies: number
  averageSalary: number
  maxSalary: number
  monthly: MonthCount[]
}

/** Mirrors dto.APIResponse — the envelope every endpoint replies with. */
export interface ApiResponse<T> {
  message: string
  data?: T
  meta?: Pagination
  error?: string
  fields?: Record<string, string>
}
