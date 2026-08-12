import { useEffect, useState, type FormEvent } from 'react'
import styles from './OpeningForm.module.css'
import { Modal } from '../ui/Modal'
import { Button } from '../ui/Button'
import { Field, Toggle } from '../ui/Field'
import { Icon } from '../ui/Icon'
import type { Opening, OpeningPayload } from '../../types/opening'

interface OpeningFormProps {
  open: boolean
  /** When set, the form edits this opening; otherwise it creates a new one. */
  editing: Opening | null
  submitting: boolean
  /**
   * Per-field messages from the API's 422, keyed by the same names this form
   * uses. They cover the rules only the server can enforce.
   */
  fieldErrors?: Record<string, string>
  onClose: () => void
  onSubmit: (payload: OpeningPayload) => void
}

type FormState = {
  role: string
  company: string
  location: string
  link: string
  salary: string
  remote: boolean
}

type FormErrors = Partial<Record<keyof FormState, string>>

const EMPTY: FormState = {
  role: '',
  company: '',
  location: '',
  link: '',
  salary: '',
  remote: false,
}

/** Mirrors the backend's validation so the user sees errors before a round-trip. */
function validate(form: FormState): FormErrors {
  const errors: FormErrors = {}

  if (!form.role.trim()) errors.role = 'Informe o cargo da vaga.'
  if (!form.company.trim()) errors.company = 'Informe a empresa.'
  if (!form.location.trim()) errors.location = 'Informe a localidade.'

  if (!form.link.trim()) {
    errors.link = 'Informe o link da vaga.'
  } else if (!/^https?:\/\/.+/i.test(form.link.trim())) {
    errors.link = 'O link precisa começar com http:// ou https://'
  }

  const salary = Number(form.salary)
  if (!form.salary.trim()) {
    errors.salary = 'Informe o salário.'
  } else if (!Number.isFinite(salary) || salary < 0) {
    errors.salary = 'O salário não pode ser negativo.'
  }

  return errors
}

export function OpeningForm({
  open,
  editing,
  submitting,
  fieldErrors,
  onClose,
  onSubmit,
}: OpeningFormProps) {
  const [form, setForm] = useState<FormState>(EMPTY)
  const [errors, setErrors] = useState<FormErrors>({})

  // What the server rejected wins over what the client checked: it is the later
  // and stricter of the two verdicts.
  const shown: FormErrors = { ...errors, ...(fieldErrors as FormErrors) }

  // Reload the fields whenever the modal opens for a different record.
  useEffect(() => {
    if (!open) return

    setErrors({})
    setForm(
      editing
        ? {
            role: editing.role,
            company: editing.company,
            location: editing.location,
            link: editing.link,
            salary: String(editing.salary),
            remote: editing.remote,
          }
        : EMPTY,
    )
  }, [open, editing])

  const patch = (partial: Partial<FormState>) => setForm((current) => ({ ...current, ...partial }))

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()

    const found = validate(form)
    setErrors(found)
    if (Object.keys(found).length > 0) return

    onSubmit({
      role: form.role.trim(),
      company: form.company.trim(),
      location: form.location.trim(),
      link: form.link.trim(),
      salary: Number(form.salary),
      remote: form.remote,
    })
  }

  return (
    <Modal
      open={open}
      size="lg"
      title={editing ? 'Editar vaga' : 'Publicar nova vaga'}
      description={
        editing
          ? 'Atualize as informações desta oportunidade.'
          : 'Preencha os dados da oportunidade que deseja divulgar.'
      }
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={submitting}>
            Cancelar
          </Button>
          <Button variant="primary" type="submit" form="opening-form" loading={submitting}>
            {editing ? 'Salvar alterações' : 'Publicar vaga'}
          </Button>
        </>
      }
    >
      <form id="opening-form" className={styles.form} onSubmit={handleSubmit} noValidate>
        <Field
          label="Cargo"
          placeholder="Ex.: Desenvolvedor Back-end Go"
          value={form.role}
          error={shown.role}
          onChange={(event) => patch({ role: event.target.value })}
          autoFocus
        />

        <div className={styles.pair}>
          <Field
            label="Empresa"
            placeholder="Ex.: Acme Corp"
            value={form.company}
            error={shown.company}
            prefix={<Icon name="building" size={15} />}
            onChange={(event) => patch({ company: event.target.value })}
          />
          <Field
            label="Localidade"
            placeholder="Ex.: São Paulo, SP"
            value={form.location}
            error={shown.location}
            prefix={<Icon name="pin" size={15} />}
            onChange={(event) => patch({ location: event.target.value })}
          />
        </div>

        <div className={styles.pair}>
          <Field
            label="Salário mensal (R$)"
            type="number"
            min={0}
            placeholder="10000"
            value={form.salary}
            error={shown.salary}
            hint="Somente números, sem pontuação. Use 0 para “a combinar”."
            prefix={<Icon name="salary" size={15} />}
            onChange={(event) => patch({ salary: event.target.value })}
          />
          <Field
            label="Link da vaga"
            type="url"
            placeholder="https://empresa.com/vagas/123"
            value={form.link}
            error={shown.link}
            prefix={<Icon name="external" size={15} />}
            onChange={(event) => patch({ link: event.target.value })}
          />
        </div>

        <Toggle
          label="Trabalho remoto"
          description="A pessoa contratada pode trabalhar de qualquer lugar."
          checked={form.remote}
          onChange={(remote) => patch({ remote })}
        />
      </form>
    </Modal>
  )
}
