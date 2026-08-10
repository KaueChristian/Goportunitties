import { describe, expect, it } from 'vitest'
import { DEFAULT_FILTERS, isFiltered, sameResultSet, toQuery, type Filters } from './filters'

function withFilters(partial: Partial<Filters>): Filters {
  return { ...DEFAULT_FILTERS, ...partial }
}

describe('isFiltered', () => {
  it('is false for the defaults', () => {
    expect(isFiltered(DEFAULT_FILTERS)).toBe(false)
  })

  // Sorting reorders the same set, so it must not light up "clear filters".
  it('ignores the sort order', () => {
    expect(isFiltered(withFilters({ sort: 'salary-desc' }))).toBe(false)
  })

  it.each<[string, Partial<Filters>]>([
    ['search', { search: 'go' }],
    ['location', { location: 'Curitiba, PR' }],
    ['modality', { remote: 'remote' }],
    ['minimum salary', { minSalary: 5000 }],
  ])('is true when %s narrows the set', (_label, partial) => {
    expect(isFiltered(withFilters(partial))).toBe(true)
  })
})

describe('toQuery', () => {
  it('maps the three-state modality onto an optional boolean', () => {
    expect(toQuery(withFilters({ remote: 'all' })).remote).toBeUndefined()
    expect(toQuery(withFilters({ remote: 'remote' })).remote).toBe(true)
    expect(toQuery(withFilters({ remote: 'onsite' })).remote).toBe(false)
  })

  it('carries the page window through', () => {
    const query = toQuery(DEFAULT_FILTERS, 3, 12)

    expect(query.page).toBe(3)
    expect(query.pageSize).toBe(12)
  })

  it('passes the location as typed — the API decides what "all" means', () => {
    expect(toQuery(withFilters({ location: 'Curitiba, PR' })).location).toBe('Curitiba, PR')
  })
})

describe('sameResultSet', () => {
  it('is true when only the sort changed', () => {
    expect(sameResultSet(DEFAULT_FILTERS, withFilters({ sort: 'role' }))).toBe(true)
  })

  it.each<[string, Partial<Filters>]>([
    ['search', { search: 'go' }],
    ['location', { location: 'Curitiba, PR' }],
    ['modality', { remote: 'onsite' }],
    ['minimum salary', { minSalary: 1000 }],
  ])('is false when %s changed', (_label, partial) => {
    expect(sameResultSet(DEFAULT_FILTERS, withFilters(partial))).toBe(false)
  })
})
