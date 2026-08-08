import { api } from './client'
import type { Opening, OpeningPayload } from '../types/opening'

/**
 * The only module that knows the shape of the openings endpoints. Components
 * and hooks talk to this, never to fetch directly.
 */
export const openingsApi = {
  list: () => api.get<Opening[]>('/openings'),

  show: (id: number) => api.get<Opening>(`/opening/${id}`),

  create: (payload: OpeningPayload) => api.post<Opening>('/opening', payload),

  update: (id: number, payload: Partial<OpeningPayload>) =>
    api.put<Opening>(`/opening/${id}`, payload),

  remove: (id: number) => api.delete<Opening>(`/opening/${id}`),
}
