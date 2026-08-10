import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { Select, type SelectOption } from './Select'

const OPTIONS: SelectOption[] = [
  { value: 'all', label: 'Todas as localidades' },
  { value: 'campinas', label: 'Campinas, SP', count: 1 },
  { value: 'curitiba', label: 'Curitiba, PR', count: 4 },
  { value: 'sao-paulo', label: 'São Paulo, SP', count: 2 },
]

function renderSelect(value = 'all') {
  const onChange = vi.fn()
  render(<Select value={value} options={OPTIONS} onChange={onChange} label="Filtrar por localidade" />)

  return { onChange, trigger: screen.getByRole('combobox', { name: 'Filtrar por localidade' }) }
}

describe('Select', () => {
  it('shows the selected option and keeps the list closed', () => {
    const { trigger } = renderSelect('curitiba')

    expect(trigger).toHaveTextContent('Curitiba, PR')
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('opens on click and reports the selected option', async () => {
    const user = userEvent.setup()
    const { trigger } = renderSelect('curitiba')

    await user.click(trigger)

    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('listbox')).toBeInTheDocument()
    expect(screen.getByRole('option', { name: /Curitiba/ })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('option', { name: /Campinas/ })).toHaveAttribute('aria-selected', 'false')
  })

  it('renders the option counts', async () => {
    const user = userEvent.setup()
    const { trigger } = renderSelect()

    await user.click(trigger)

    expect(screen.getByRole('option', { name: /Curitiba/ })).toHaveTextContent('4')
  })

  it('reports the chosen value and closes', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    await user.click(trigger)
    await user.click(screen.getByRole('option', { name: /São Paulo/ }))

    expect(onChange).toHaveBeenCalledWith('sao-paulo')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('walks the options with the arrow keys and picks with Enter', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    trigger.focus()
    await user.keyboard('{ArrowDown}') // opens on the selected option
    await user.keyboard('{ArrowDown}{ArrowDown}')
    await user.keyboard('{Enter}')

    expect(onChange).toHaveBeenCalledWith('curitiba')
  })

  it('stops at the ends instead of wrapping around', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    trigger.focus()
    await user.keyboard('{ArrowDown}')
    await user.keyboard('{ArrowUp}{ArrowUp}{ArrowUp}')
    await user.keyboard('{Enter}')

    expect(onChange).toHaveBeenCalledWith('all')
  })

  it('jumps to the first and last option with Home and End', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    trigger.focus()
    await user.keyboard('{ArrowDown}{End}{Enter}')

    expect(onChange).toHaveBeenCalledWith('sao-paulo')
  })

  it('jumps to an option by its first letter', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    trigger.focus()
    await user.keyboard('{ArrowDown}')
    await user.keyboard('c') // Campinas
    await user.keyboard('{Enter}')

    expect(onChange).toHaveBeenCalledWith('campinas')
  })

  it('closes on Escape without choosing', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    await user.click(trigger)
    await user.keyboard('{ArrowDown}')
    await user.keyboard('{Escape}')

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(onChange).not.toHaveBeenCalled()
  })

  it('closes when the pointer goes elsewhere', async () => {
    const user = userEvent.setup()
    const { trigger, onChange } = renderSelect()

    await user.click(trigger)
    await user.click(document.body)

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(onChange).not.toHaveBeenCalled()
  })

  // A value with no matching option must not blank the trigger.
  it('falls back to the first option when the value is unknown', () => {
    const { trigger } = renderSelect('inexistente')

    expect(trigger).toHaveTextContent('Todas as localidades')
  })
})
