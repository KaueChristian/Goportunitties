import type { ApiResponse } from '../types/opening'

/**
 * Base URL of the API. In development Vite proxies /api to the Go server, so a
 * relative path keeps the browser on a single origin. Override with VITE_API_URL
 * when the frontend is deployed separately from the backend.
 */
const BASE_URL = import.meta.env.VITE_API_URL ?? '/api/v1'

/** An error carrying the HTTP status, so callers can branch on 404 vs 500. */
export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }

  /** True when the request never reached the server (offline, API down). */
  get isNetworkError(): boolean {
    return this.status === 0
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
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
    )
  }

  return body?.data as T
}

export const api = {
  get: <T>(path: string) => request<T>(path, { method: 'GET' }),

  post: <T>(path: string, payload: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(payload) }),

  put: <T>(path: string, payload: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(payload) }),

  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
