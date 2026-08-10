import type { ApiResponse, Pagination } from '../types/opening'

/**
 * Base URL of the API. In development Vite proxies /api to the Go server, and in
 * production the same binary serves both — so a relative path is right in both
 * cases. VITE_API_URL covers deploying the frontend somewhere else.
 */
const BASE_URL = import.meta.env.VITE_API_URL ?? '/api/v1'

/** An error carrying the HTTP status, so callers can branch on 404 vs 500. */
export class ApiError extends Error {
  readonly status: number
  /** Per-field messages from a 422, keyed by the field name the API uses. */
  readonly fields: Record<string, string>

  constructor(message: string, status: number, fields: Record<string, string> = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fields = fields
  }

  /** True when the request never reached the server (offline, API down). */
  get isNetworkError(): boolean {
    return this.status === 0
  }

  /** True when the server rejected specific fields rather than the request. */
  get isValidationError(): boolean {
    return this.status === 422
  }
}

/** A response body split into its payload and, for collections, its window. */
export interface ApiResult<T> {
  data: T
  meta?: Pagination
}

/** Values a query string can carry; nullish entries are dropped. */
export type QueryParams = Record<string, string | number | boolean | null | undefined>

/** Builds a query string, omitting anything the caller left unset. */
export function toQueryString(params: QueryParams): string {
  const search = new URLSearchParams()

  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined || value === '') continue
    search.set(key, String(value))
  }

  const query = search.toString()
  return query ? `?${query}` : ''
}

async function request<T>(path: string, init: RequestInit = {}): Promise<ApiResult<T>> {
  let response: Response

  try {
    response = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...init.headers,
      },
    })
  } catch {
    throw new ApiError(
      'Não foi possível conectar à API. Verifique se o servidor Go está rodando.',
      0,
    )
  }

  let body: ApiResponse<T> | null = null
  try {
    body = (await response.json()) as ApiResponse<T>
  } catch {
    // A body that isn't JSON (e.g. a proxy error page) is handled below.
  }

  if (!response.ok) {
    throw new ApiError(
      body?.error ?? `A requisição falhou com status ${response.status}.`,
      response.status,
      body?.fields ?? {},
    )
  }

  return { data: body?.data as T, meta: body?.meta }
}

export const api = {
  get: <T>(path: string) => request<T>(path, { method: 'GET' }),

  post: <T>(path: string, payload: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(payload) }),

  put: <T>(path: string, payload: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(payload) }),

  patch: <T>(path: string, payload: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(payload) }),

  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
