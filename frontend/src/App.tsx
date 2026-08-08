import { useMemo, useState } from 'react'
import styles from './App.module.css'

import { Header } from './components/layout/Header'
import { Footer } from './components/layout/Footer'
import { HomePage } from './pages/HomePage'
import { JobsPage } from './pages/JobsPage'
import { DashboardPage } from './pages/DashboardPage'
import { OpeningForm } from './components/openings/OpeningForm'
import { ConfirmDialog } from './components/ui/ConfirmDialog'
import { useToast } from './components/ui/Toast'

import { DEFAULT_FILTERS, type Filters } from './components/openings/filters'
import { useOpenings } from './hooks/useOpenings'
import { useTheme } from './hooks/useTheme'
import { useHashRoute } from './router/useHashRoute'
import { ApiError } from './api/client'
import type { Opening, OpeningPayload } from './types/opening'

export default function App() {
  const { theme, toggleTheme } = useTheme()
  const { route, navigate } = useHashRoute()
  const { openings, loading, error, reload, create, update, remove } = useOpenings()
  const { notifySuccess, notifyError } = useToast()

  /*
   * Filters and selection live here rather than in JobsPage: the hero on the
   * home page seeds a search and hands the user over to the results, and both
   * other pages can open a card's detail on the jobs page.
   */
  const [filters, setFilters] = useState<Filters>(DEFAULT_FILTERS)
  const [selected, setSelected] = useState<Opening | null>(null)

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Opening | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [pendingDelete, setPendingDelete] = useState<Opening | null>(null)
  const [deleting, setDeleting] = useState(false)

  const locations = useMemo(
    () =>
      [...new Set(openings.map((opening) => opening.location))].sort((a, b) =>
        a.localeCompare(b),
      ),
    [openings],
  )

  const openCreate = () => {
    setEditing(null)
    setFormOpen(true)
  }

  const openEdit = (opening: Opening) => {
    setEditing(opening)
    setFormOpen(true)
  }

  const openDetail = (opening: Opening) => {
    setSelected(opening)
    navigate('jobs')
  }

  const handleSubmit = async (payload: OpeningPayload) => {
    setSubmitting(true)
    try {
      if (editing) {
        const updated = await update(editing.id, payload)
        setSelected((current) => (current?.id === updated.id ? updated : current))
        notifySuccess('Vaga atualizada com sucesso.')
      } else {
        await create(payload)
        notifySuccess('Vaga publicada com sucesso.')
      }
      setFormOpen(false)
      setEditing(null)
    } catch (err) {
      notifyError(err instanceof ApiError ? err.message : 'Não foi possível salvar a vaga.')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!pendingDelete) return

    setDeleting(true)
    try {
      await remove(pendingDelete.id)
      setSelected((current) => (current?.id === pendingDelete.id ? null : current))
      notifySuccess('Vaga excluída.')
      setPendingDelete(null)
    } catch (err) {
      notifyError(err instanceof ApiError ? err.message : 'Não foi possível excluir a vaga.')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className={styles.app}>
      <Header route={route} theme={theme} onToggleTheme={toggleTheme} onCreate={openCreate} />

      <main className={styles.main}>
        {route === 'home' && (
          <HomePage
            openings={openings}
            loading={loading}
            filters={filters}
            locations={locations}
            onFiltersChange={setFilters}
            onSearch={() => navigate('jobs')}
            onSelect={openDetail}
            onEdit={openEdit}
            onDelete={setPendingDelete}
            onCreate={openCreate}
          />
        )}

        {route === 'jobs' && (
          <JobsPage
            openings={openings}
            loading={loading}
            error={error}
            filters={filters}
            locations={locations}
            selected={selected}
            onFiltersChange={setFilters}
            onSelect={setSelected}
            onCloseDetail={() => setSelected(null)}
            onEdit={openEdit}
            onDelete={setPendingDelete}
            onCreate={openCreate}
            onReload={() => void reload()}
          />
        )}

        {route === 'dashboard' && (
          <DashboardPage
            openings={openings}
            loading={loading}
            error={error}
            onSelect={openDetail}
            onEdit={openEdit}
            onDelete={setPendingDelete}
            onCreate={openCreate}
            onReload={() => void reload()}
          />
        )}
      </main>

      <Footer />

      <OpeningForm
        open={formOpen}
        editing={editing}
        submitting={submitting}
        onClose={() => {
          setFormOpen(false)
          setEditing(null)
        }}
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
    </div>
  )
}
