import { describe, expect, it } from 'vitest'
import { isEditable, isManual, sourceLabel } from './sources'
import type { Opening } from '../../types/opening'

function opening(source: string): Opening {
  return {
    id: 1,
    createdAt: '2026-08-01T12:00:00Z',
    updatedAt: '2026-08-01T12:00:00Z',
    source,
    role: 'Desenvolvedor Go',
    company: 'Acme',
    location: 'Remoto',
    remote: true,
    link: 'https://acme.com/1',
    salary: 0,
  }
}

describe('sourceLabel', () => {
  it.each([
    ['manual', 'Publicada aqui'],
    ['remoteok', 'RemoteOK'],
    ['remotive', 'Remotive'],
  ])('spells %o as %o', (slug, expected) => {
    expect(sourceLabel(slug)).toBe(expected)
  })

  // A board added to the worker keeps working here before anyone updates the
  // map, which is better than showing a placeholder.
  it('falls back to the capitalised slug', () => {
    expect(sourceLabel('weworkremotely')).toBe('Weworkremotely')
  })
})

describe('isManual', () => {
  it('separates typed openings from ingested ones', () => {
    expect(isManual(opening('manual'))).toBe(true)
    expect(isManual(opening('remoteok'))).toBe(false)
  })
})

describe('isEditable', () => {
  it('allows editing what was typed here', () => {
    expect(isEditable(opening('manual'))).toBe(true)
  })

  // The board owns the content: the next ingestion run would undo the edit.
  it('refuses to edit an ingested opening', () => {
    expect(isEditable(opening('remoteok'))).toBe(false)
  })
})
