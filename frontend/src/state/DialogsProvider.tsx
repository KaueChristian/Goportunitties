import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'

import { OpeningForm } from '../components/openings/OpeningForm'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { useToast } from '../components/ui/Toast'
import { useOpenings } from './OpeningsProvider'
import { ApiError } from '../api/client'
import type { Opening, OpeningPayload } from '../types/opening'

interface DialogsContextValue {
  openCreate: () => void
  openEdit: (opening: Opening) => void
  confirmDelete: (opening: Opening) => void
}

const DialogsContext = createContext<DialogsContextValue | null>(null)

/**
 * Owns the two modals the whole app can raise: the opening form and the delete
 * confirmation. Keeping them here means a card deep in a page can ask for one
 * without its parents having to thread `onEdit`/`onDelete` down to it.
 */
export function DialogsProvider({ children }: { children: ReactNode }) {
  const { createOpening, updateOpening, deleteOpening } = useOpenings()
  const { notifySuccess, notifyError } = useToast()

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Opening | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  const [pendingDelete, setPendingDelete] = useState<Opening | null>(null)
  const [deleting, setDeleting] = useState(false)

  const openCreate = useCallback(() => {
    setEditing(null)
    setFieldErrors({})
    setFormOpen(true)
  }, [])

  const openEdit = useCallback((opening: Opening) => {
    setEditing(opening)
    setFieldErrors({})
    setFormOpen(true)
  }, [])

  const closeForm = useCallback(() => {
    setFormOpen(false)
    setEditing(null)
    setFieldErrors({})
  }, [])

  const handleSubmit = useCallback(
    async (payload: OpeningPayload) => {
      setSubmitting(true)
      setFieldErrors({})

      try {
        if (editing) {
          await updateOpening(editing.id, payload)
          notifySuccess('Vaga atualizada com sucesso.')
        } else {
          await createOpening(payload)
          notifySuccess('Vaga publicada com sucesso.')
        }
        closeForm()
      } catch (err) {
        // A 422 names the offending fields, so the form can point at them
        // instead of showing one vague toast.
        if (err instanceof ApiError && err.isValidationError) {
          setFieldErrors(err.fields)
          notifyError('Confira os campos destacados no formulário.')
          return
        }
        notifyError(err instanceof ApiError ? err.message : 'Não foi possível salvar a vaga.')
      } finally {
        setSubmitting(false)
      }
    },
    [editing, createOpening, updateOpening, notifySuccess, notifyError, closeForm],
  )

  const handleDelete = useCallback(async () => {
    if (!pendingDelete) return

    setDeleting(true)
    try {
      await deleteOpening(pendingDelete.id)
      notifySuccess('Vaga excluída.')
      setPendingDelete(null)
    } catch (err) {
      notifyError(err instanceof ApiError ? err.message : 'Não foi possível excluir a vaga.')
    } finally {
      setDeleting(false)
    }
  }, [pendingDelete, deleteOpening, notifySuccess, notifyError])

  const value = useMemo<DialogsContextValue>(
    () => ({ openCreate, openEdit, confirmDelete: setPendingDelete }),
    [openCreate, openEdit],
  )

  return (
    <DialogsContext.Provider value={value}>
      {children}

      <OpeningForm
        open={formOpen}
        editing={editing}
        submitting={submitting}
        fieldErrors={fieldErrors}
        onClose={closeForm}
        onSubmit={handleSubmit}
      />

      <ConfirmDialog
        open={pendingDelete !== null}
        title="Excluir vaga"
        message={
          pendingDelete
            ? `Tem certeza que deseja excluir a vaga "${pendingDelete.role}" na ${pendingDelete.company}? Essa ação não pode ser desfeita.`
            : ''
        }
        confirmLabel="Excluir vaga"
        loading={deleting}
        onConfirm={() => void handleDelete()}
        onCancel={() => setPendingDelete(null)}
      />
    </DialogsContext.Provider>
  )
}

export function useDialogs(): DialogsContextValue {
  const context = useContext(DialogsContext)
  if (!context) {
    throw new Error('useDialogs precisa estar dentro de <DialogsProvider>')
  }
  return context
}
