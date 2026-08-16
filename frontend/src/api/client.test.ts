import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, api, toQueryString } from './client'

/** Replaces global fetch with one that answers `body` under `status`. */
function mockFetch(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  })
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
  localStorage.clear()
})

describe('toQueryString', () => {
  it('drops everything the caller left unset', () => {
    const query = toQueryString({
      search: 'go',
      location: undefined,
      remote: null,
      minSalary: 0,
      page: 2,
      empty: '',
    })

    // `minSalary: 0` survives because zero is a value; '' and nullish do not.
    expect(query).toBe('?search=go&minSalary=0&page=2')
  })

  it('returns an empty string when there is nothing to send', () => {
    expect(toQueryString({ search: undefined })).toBe('')
  })

  it('escapes values', () => {
    expect(toQueryString({ location: 'São Paulo, SP' })).toContain('S%C3%A3o+Paulo%2C+SP')
  })
})

describe('api', () => {
  it('unwraps the envelope and keeps the pagination meta', async () => {
    mockFetch(200, {
      message: 'list-openings successful',
      data: [{ id: 1 }],
      meta: { page: 1, pageSize: 12, total: 57, totalPages: 5 },
    })

    const result = await api.get<{ id: number }[]>('/openings')

    expect(result.data).toEqual([{ id: 1 }])
    expect(result.meta?.total).toBe(57)
  })

  it('sends the payload as JSON', async () => {
    const fetchMock = mockFetch(201, { message: 'ok', data: { id: 1 } })

    await api.post('/openings', { role: 'SRE' })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/openings')
    expect(init.method).toBe('POST')
    expect(init.body).toBe('{"role":"SRE"}')
    expect(init.headers['Content-Type']).toBe('application/json')
  })

  it('turns a 404 into an ApiError carrying the status', async () => {
    mockFetch(404, { message: 'operation failed', error: 'vaga não encontrada' })

    const error = await api.get('/openings/9').catch((err: unknown) => err)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(404)
    expect((error as ApiError).message).toBe('vaga não encontrada')
    expect((error as ApiError).isValidationError).toBe(false)
  })

  it('exposes the per-field messages of a 422', async () => {
    mockFetch(422, {
      message: 'validation failed',
      error: 'há campos inválidos na requisição',
      fields: { link: 'informe uma URL começando com http:// ou https://' },
    })

    const error = (await api.post('/openings', {}).catch((err: unknown) => err)) as ApiError

    expect(error.isValidationError).toBe(true)
    expect(error.fields.link).toContain('http://')
  })

  it('reports an unreachable API instead of throwing the raw fetch failure', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    const error = (await api.get('/openings').catch((err: unknown) => err)) as ApiError

    expect(error.isNetworkError).toBe(true)
    expect(error.message).toContain('servidor Go')
  })

  it('reports a blocked write distinctly from other failures', async () => {
    mockFetch(401, { message: 'operation failed', error: 'chave de administrador ausente ou inválida' })

    const error = (await api.post('/openings', {}).catch((err: unknown) => err)) as ApiError

    expect(error.isUnauthorized).toBe(true)
    expect(error.isValidationError).toBe(false)
  })

  it('attaches the stored admin key to a write', async () => {
    localStorage.setItem('goportunitties:admin-key', 's3cr3t')
    const fetchMock = mockFetch(201, { message: 'ok', data: {} })

    await api.post('/openings', { role: 'SRE' })

    expect(fetchMock.mock.calls[0][1].headers['X-Admin-Key']).toBe('s3cr3t')
  })

  it('sends no admin key header on a write when none is stored', async () => {
    const fetchMock = mockFetch(201, { message: 'ok', data: {} })

    await api.post('/openings', { role: 'SRE' })

    expect(fetchMock.mock.calls[0][1].headers['X-Admin-Key']).toBeUndefined()
  })

  it('survives an error response that is not JSON', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 502,
        json: async () => {
          throw new SyntaxError('Unexpected token <')
        },
      }),
    )

    const error = (await api.get('/openings').catch((err: unknown) => err)) as ApiError

    expect(error.status).toBe(502)
    expect(error.message).toContain('502')
  })
})
