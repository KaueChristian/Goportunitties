import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { avatarColor, formatRelativeDate, formatSalary, initialsOf, normalize } from './format'

describe('formatSalary', () => {
  it('formats as Brazilian currency without cents', () => {
    // The separators Intl emits are non-breaking spaces, so compare loosely.
    const formatted = formatSalary(15000).replace(/ /g, ' ')

    expect(formatted).toContain('R$')
    expect(formatted).toContain('15.000')
    expect(formatted).not.toContain(',00')
  })

  it('formats zero rather than blanking it', () => {
    expect(formatSalary(0)).toContain('0')
  })
})

describe('initialsOf', () => {
  it.each([
    ['Acme Corp', 'AC'],
    ['Globex', 'GL'],
    ['  Initech  ', 'IN'],
    ['a', 'A'],
    ['', '?'],
    ['   ', '?'],
  ])('turns %o into %o', (name, expected) => {
    expect(initialsOf(name)).toBe(expected)
  })

  // Strictly the first two words, connectives included — the avatar is a
  // recognition aid, not an acronym generator.
  it('uses the first letter of the first two words, whatever they are', () => {
    expect(initialsOf('Red Hat Brasil')).toBe('RH')
    expect(initialsOf('Banco do Brasil')).toBe('BD')
  })
})

describe('avatarColor', () => {
  it('is stable for the same company', () => {
    expect(avatarColor('Acme')).toBe(avatarColor('Acme'))
  })

  it('picks from the fixed palette', () => {
    const color = avatarColor('Qualquer Empresa Ltda')

    expect(color).toMatch(/^(#[0-9a-f]{6}|var\(--[a-z0-9-]+\))$/i)
  })
})

describe('normalize', () => {
  it('strips case and accents so search matches either spelling', () => {
    expect(normalize('São Paulo')).toBe('sao paulo')
    expect(normalize('DESENVOLVEDOR')).toBe('desenvolvedor')
    expect(normalize('Análise')).toBe(normalize('analise'))
  })
})

describe('formatRelativeDate', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-08-15T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('says "hoje" for today', () => {
    expect(formatRelativeDate('2026-08-15T09:00:00Z')).toBe('hoje')
  })

  it('counts in days within the last month', () => {
    expect(formatRelativeDate('2026-08-12T12:00:00Z')).toContain('3 dias')
  })

  it('switches to months further back', () => {
    expect(formatRelativeDate('2026-05-15T12:00:00Z')).toContain('meses')
  })

  it('switches to years past twelve months', () => {
    expect(formatRelativeDate('2024-08-15T12:00:00Z')).toContain('ano')
  })
})
