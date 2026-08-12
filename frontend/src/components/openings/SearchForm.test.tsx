import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { SearchForm } from './SearchForm'
import { DEFAULT_FILTERS, type Filters } from './filters'
import type { Suggestion } from '../../types/opening'

const SUGGESTIONS: Suggestion[] = [
  { value: 'Desenvolvedor Go', kind: 'role', count: 3 },
  { value: 'Globex', kind: 'company', count: 1 },
]

let fetchMock: ReturnType<typeof vi.fn>

function mockSuggestions(data: Suggestion[]) {
  fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => ({ message: 'ok', data }),
  })
  vi.stubGlobal('fetch', fetchMock)
}

beforeEach(() => {
  mockSuggestions(SUGGESTIONS)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function renderForm(filters: Filters = DEFAULT_FILTERS) {
  const onChange = vi.fn()
  const onSubmit = vi.fn()

  render(
    <SearchForm
      filters={filters}
      locations={['Campinas, SP', 'Curitiba, PR']}
      onChange={onChange}
      onSubmit={onSubmit}
    />,
  )

  return { onChange, onSubmit, input: screen.getByRole('combobox', { name: /Buscar por cargo/ }) }
}

describe('SearchForm autocomplete', () => {
  it('suggests roles and companies once the term is long enough', async () => {
    const user = userEvent.setup()
    const { input } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)

    expect(await screen.findByRole('option', { name: /Desenvolvedor Go/ })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: /Globex/ })).toBeInTheDocument()
    expect(input).toHaveAttribute('aria-expanded', 'true')
  })

  it('labels each suggestion with how many openings carry it', async () => {
    const user = userEvent.setup()
    const { input } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)

    expect(await screen.findByRole('option', { name: /Desenvolvedor Go/ })).toHaveTextContent('3 vagas')
    expect(screen.getByRole('option', { name: /Globex/ })).toHaveTextContent('1 vaga')
  })

  // One letter matches nearly everything; the request is not worth making.
  it('asks for nothing while the term is a single character', async () => {
    const user = userEvent.setup()
    const { input } = renderForm({ ...DEFAULT_FILTERS, search: 'g' })

    await user.click(input)

    await new Promise((resolve) => setTimeout(resolve, 300))
    expect(fetchMock).not.toHaveBeenCalled()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('applies a suggestion and runs the search', async () => {
    const user = userEvent.setup()
    const { input, onChange, onSubmit } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)
    await user.click(await screen.findByRole('option', { name: /Globex/ }))

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ search: 'Globex' }))
    expect(onSubmit).toHaveBeenCalled()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('walks the suggestions with the arrow keys and applies with Enter', async () => {
    const user = userEvent.setup()
    const { input, onChange } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)
    await screen.findByRole('option', { name: /Desenvolvedor Go/ })

    await user.keyboard('{ArrowDown}{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ search: 'Globex' }))
  })

  // Enter with nothing highlighted submits what was typed, as any search box does.
  it('submits the typed term when no suggestion is highlighted', async () => {
    const user = userEvent.setup()
    const { input, onSubmit, onChange } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)
    await screen.findByRole('option', { name: /Desenvolvedor Go/ })

    await user.keyboard('{Enter}')

    expect(onSubmit).toHaveBeenCalled()
    expect(onChange).not.toHaveBeenCalled()
  })

  it('closes the list on Escape', async () => {
    const user = userEvent.setup()
    const { input } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)
    await screen.findByRole('option', { name: /Globex/ })

    await user.keyboard('{Escape}')

    await waitFor(() => expect(screen.queryByRole('listbox')).not.toBeInTheDocument())
  })

  it('stays usable when the suggestions request fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    const user = userEvent.setup()
    const { input, onSubmit } = renderForm({ ...DEFAULT_FILTERS, search: 'go' })

    await user.click(input)
    await new Promise((resolve) => setTimeout(resolve, 300))

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    await user.keyboard('{Enter}')
    expect(onSubmit).toHaveBeenCalled()
  })
})

describe('SearchForm filters', () => {
  it('offers the locations it was given', async () => {
    const user = userEvent.setup()
    renderForm()

    await user.click(screen.getByRole('combobox', { name: 'Filtrar por localidade' }))

    expect(screen.getByRole('option', { name: 'Curitiba, PR' })).toBeInTheDocument()
  })

  it('reports the chosen modality', async () => {
    const user = userEvent.setup()
    const { onChange } = renderForm()

    await user.click(screen.getByRole('combobox', { name: 'Filtrar por modalidade' }))
    await user.click(screen.getByRole('option', { name: 'Presencial' }))

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ remote: 'onsite' }))
  })
})
