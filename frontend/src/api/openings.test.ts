import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { openingsApi } from './openings'

let fetchMock: ReturnType<typeof vi.fn>

beforeEach(() => {
  fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => ({ message: 'ok', data: [], meta: undefined }),
  })
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function requestedUrl(): string {
  return fetchMock.mock.calls[0][0] as string
}

describe('openingsApi.list', () => {
  it('translates a full query into parameters', async () => {
    await openingsApi.list({
      search: '  go  ',
      location: 'Curitiba, PR',
      remote: true,
      minSalary: 8000,
      sort: 'salary-desc',
      page: 2,
      pageSize: 12,
    })

    const url = requestedUrl()
    expect(url).toContain('search=go')
    expect(url).toContain('location=Curitiba%2C+PR')
    expect(url).toContain('remote=true')
    expect(url).toContain('minSalary=8000')
    expect(url).toContain('sort=salary-desc')
    expect(url).toContain('page=2&pageSize=12')
  })

  // 'all' and 0 are how the UI spells "no filter" — they must not reach the API.
  it('omits the filters that mean "no filter"', async () => {
    await openingsApi.list({ location: 'all', minSalary: 0, search: '   ' })

    const url = requestedUrl()
    expect(url).not.toContain('location=')
    expect(url).not.toContain('minSalary=')
    expect(url).not.toContain('search=')
  })

  it('keeps remote=false, which is a filter and not an absence', async () => {
    await openingsApi.list({ remote: false })

    expect(requestedUrl()).toContain('remote=false')
  })

  it('falls back to a single-page window when the response carries no meta', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ message: 'ok', data: [{ id: 1 }, { id: 2 }] }),
    })

    const page = await openingsApi.list()

    expect(page.openings).toHaveLength(2)
    expect(page.pagination).toEqual({ page: 1, pageSize: 2, total: 2, totalPages: 1 })
  })

  it('returns an empty array when data is missing', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ message: 'ok' }),
    })

    const page = await openingsApi.list()

    expect(page.openings).toEqual([])
  })
})

describe('openingsApi writes', () => {
  it('uses PUT for a full replacement and PATCH for a partial one', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ message: 'ok', data: { id: 7 } }),
    })

    await openingsApi.replace(7, {
      role: 'SRE',
      company: 'Initech',
      location: 'Curitiba, PR',
      remote: false,
      link: 'https://initech.com/9',
      salary: 21500,
    })
    expect(fetchMock.mock.calls[0][1].method).toBe('PUT')

    await openingsApi.patch(7, { salary: 0 })
    expect(fetchMock.mock.calls[1][1].method).toBe('PATCH')
    expect(fetchMock.mock.calls[1][1].body).toBe('{"salary":0}')
  })
})
