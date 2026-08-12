import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { OpeningCard } from './OpeningCard'
import type { Opening } from '../../types/opening'

function opening(overrides: Partial<Opening> = {}): Opening {
  return {
    id: 1,
    createdAt: '2026-08-01T12:00:00Z',
    updatedAt: '2026-08-01T12:00:00Z',
    source: 'manual',
    role: 'Desenvolvedor Go',
    company: 'Acme',
    location: 'São Paulo, SP',
    remote: true,
    link: 'https://acme.com/vagas/1',
    salary: 15000,
    ...overrides,
  }
}

function renderCard(overrides: Partial<Opening> = {}) {
  const onEdit = vi.fn()
  const onDelete = vi.fn()
  const onSelect = vi.fn()

  render(
    <OpeningCard
      opening={opening(overrides)}
      index={0}
      selected={false}
      onSelect={onSelect}
      onEdit={onEdit}
      onDelete={onDelete}
    />,
  )

  return { onEdit, onDelete, onSelect }
}

describe('OpeningCard provenance', () => {
  it('says nothing about origin for an opening published here', () => {
    renderCard()

    expect(screen.queryByText(/^via /)).not.toBeInTheDocument()
  })

  it('credits the board an ingested opening came from', () => {
    renderCard({ source: 'remoteok' })

    expect(screen.getByText(/via RemoteOK/)).toBeInTheDocument()
  })

  it('offers editing for an opening published here', () => {
    renderCard()

    expect(screen.getByRole('button', { name: /Editar vaga/ })).toBeInTheDocument()
  })

  // The next ingestion run would overwrite the edit, so the control is not
  // offered at all rather than offered and then silently undone.
  it('does not offer editing for an ingested opening', () => {
    renderCard({ source: 'remoteok' })

    expect(screen.queryByRole('button', { name: /Editar vaga/ })).not.toBeInTheDocument()
  })

  // Dismissing an ingested opening is allowed: the upsert leaves it deleted.
  it('still offers deleting for an ingested opening', async () => {
    const user = userEvent.setup()
    const { onDelete } = renderCard({ source: 'remoteok' })

    await user.click(screen.getByRole('button', { name: /Excluir vaga/ }))

    expect(onDelete).toHaveBeenCalled()
  })
})

describe('OpeningCard salary', () => {
  it('shows the figure with its unit when there is one', () => {
    renderCard({ salary: 15000 })

    expect(screen.getByText('/mês')).toBeInTheDocument()
  })

  // Ingested openings arrive without a salary this application can express.
  // "R$ 0/mês" would state something the board never said.
  it('reads "A combinar" when no salary was stated', () => {
    renderCard({ salary: 0, source: 'remoteok' })

    expect(screen.getByText('A combinar')).toBeInTheDocument()
    expect(screen.queryByText('/mês')).not.toBeInTheDocument()
  })
})
