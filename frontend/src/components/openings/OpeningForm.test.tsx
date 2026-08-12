import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { OpeningForm } from './OpeningForm'
import type { Opening } from '../../types/opening'

const opening: Opening = {
  id: 7,
  createdAt: '2026-08-01T12:00:00Z',
  updatedAt: '2026-08-01T12:00:00Z',
  source: 'manual',
  role: 'Desenvolvedor Go',
  company: 'Acme',
  location: 'São Paulo, SP',
  remote: true,
  link: 'https://acme.com/vagas/1',
  salary: 15000,
}

function renderForm(props: Partial<React.ComponentProps<typeof OpeningForm>> = {}) {
  const onSubmit = vi.fn()
  const onClose = vi.fn()

  render(
    <OpeningForm
      open
      editing={null}
      submitting={false}
      onClose={onClose}
      onSubmit={onSubmit}
      {...props}
    />,
  )

  return { onSubmit, onClose }
}

/** Fills every field with something valid. */
async function fillValidForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('Cargo'), 'Desenvolvedor Go')
  await user.type(screen.getByLabelText('Empresa'), 'Acme')
  await user.type(screen.getByLabelText('Localidade'), 'São Paulo, SP')
  await user.type(screen.getByLabelText('Link da vaga'), 'https://acme.com/vagas/1')
  await user.type(screen.getByLabelText('Salário mensal (R$)'), '15000')
}

describe('OpeningForm', () => {
  it('renders nothing while closed', () => {
    renderForm({ open: false })

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('blocks an empty submit and reports every missing field', async () => {
    const user = userEvent.setup()
    const { onSubmit } = renderForm()

    await user.click(screen.getByRole('button', { name: 'Publicar vaga' }))

    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByText('Informe o cargo da vaga.')).toBeInTheDocument()
    expect(screen.getByText('Informe a empresa.')).toBeInTheDocument()
    expect(screen.getByText('Informe a localidade.')).toBeInTheDocument()
    expect(screen.getByText('Informe o link da vaga.')).toBeInTheDocument()
  })

  // Mirrors the API's httpurl rule: this value ends up in an anchor's href.
  it('rejects a link that is not http(s)', async () => {
    const user = userEvent.setup()
    const { onSubmit } = renderForm()

    await fillValidForm(user)
    await user.clear(screen.getByLabelText('Link da vaga'))
    await user.type(screen.getByLabelText('Link da vaga'), 'javascript:alert(1)')
    await user.click(screen.getByRole('button', { name: 'Publicar vaga' }))

    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByText(/http:\/\/ ou https:\/\//)).toBeInTheDocument()
  })

  it('submits the trimmed payload', async () => {
    const user = userEvent.setup()
    const { onSubmit } = renderForm()

    await user.type(screen.getByLabelText('Cargo'), '  Desenvolvedor Go  ')
    await user.type(screen.getByLabelText('Empresa'), 'Acme')
    await user.type(screen.getByLabelText('Localidade'), 'São Paulo, SP')
    await user.type(screen.getByLabelText('Link da vaga'), 'https://acme.com/vagas/1')
    await user.type(screen.getByLabelText('Salário mensal (R$)'), '15000')
    await user.click(screen.getByRole('button', { name: 'Publicar vaga' }))

    expect(onSubmit).toHaveBeenCalledWith({
      role: 'Desenvolvedor Go',
      company: 'Acme',
      location: 'São Paulo, SP',
      link: 'https://acme.com/vagas/1',
      salary: 15000,
      remote: false,
    })
  })

  // The API accepts 0 as "a combinar"; the form must not be stricter.
  it('accepts a salary of zero', async () => {
    const user = userEvent.setup()
    const { onSubmit } = renderForm()

    await fillValidForm(user)
    await user.clear(screen.getByLabelText('Salário mensal (R$)'))
    await user.type(screen.getByLabelText('Salário mensal (R$)'), '0')
    await user.click(screen.getByRole('button', { name: 'Publicar vaga' }))

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ salary: 0 }))
  })

  it('loads the opening being edited', () => {
    renderForm({ editing: opening })

    expect(screen.getByLabelText('Cargo')).toHaveValue('Desenvolvedor Go')
    expect(screen.getByLabelText('Salário mensal (R$)')).toHaveValue(15000)
    expect(screen.getByRole('switch', { name: 'Trabalho remoto' })).toHaveAttribute(
      'aria-checked',
      'true',
    )
    expect(screen.getByRole('button', { name: 'Salvar alterações' })).toBeInTheDocument()
  })

  it('shows the per-field messages the API returned', () => {
    renderForm({
      fieldErrors: {
        link: 'informe uma URL começando com http:// ou https://',
        role: 'deve ter ao menos 2 caracteres',
      },
    })

    expect(screen.getByText('deve ter ao menos 2 caracteres')).toBeInTheDocument()
    expect(screen.getByLabelText('Link da vaga')).toHaveAttribute('aria-invalid', 'true')
  })

  it('disables cancelling while a submit is in flight', () => {
    renderForm({ submitting: true })

    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeDisabled()
  })
})
