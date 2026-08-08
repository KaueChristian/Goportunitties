import { useCallback, useEffect, useState } from 'react'
import { openingsApi } from '../api/openings'
import { ApiError } from '../api/client'
import type { Opening, OpeningPayload } from '../types/opening'

interface State {
  openings: Opening[]
  loading: boolean
  error: string | null
}

/**
 * Owns the openings collection: fetching, and the write operations that keep
 * local state in sync with the server's response (no blind refetch after each
 * mutation — the API already returns the persisted record).
 */
export function useOpenings() {
  const [state, setState] = useState<State>({
    openings: [],
    loading: true,
    error: null,
  })

  const load = useCallback(async () => {
    setState((current) => ({ ...current, loading: true, error: null }))
    try {
      const openings = await openingsApi.list()
      setState({ openings: openings ?? [], loading: false, error: null })
    } catch (error) {
      setState({
        openings: [],
        loading: false,
        error: error instanceof ApiError ? error.message : 'Erro inesperado ao carregar vagas.',
      })
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const create = useCallback(async (payload: OpeningPayload) => {
    const created = await openingsApi.create(payload)
    setState((current) => ({ ...current, openings: [created, ...current.openings] }))
    return created
  }, [])

  const update = useCallback(async (id: number, payload: Partial<OpeningPayload>) => {
    const updated = await openingsApi.update(id, payload)
    setState((current) => ({
      ...current,
      openings: current.openings.map((opening) => (opening.id === id ? updated : opening)),
    }))
    return updated
  }, [])

  const remove = useCallback(async (id: number) => {
    await openingsApi.remove(id)
    setState((current) => ({
      ...current,
      openings: current.openings.filter((opening) => opening.id !== id),
    }))
  }, [])

  return { ...state, reload: load, create, update, remove }
}
