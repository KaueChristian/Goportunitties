import type { Opening } from '../../types/opening'

/** How the interface spells each source out. Mirrors the slugs in internal/ingestion. */
const SOURCE_LABELS: Record<string, string> = {
  manual: 'Publicada aqui',
  'backend-br': 'Vagas Back-end BR',
  'frontend-br': 'Vagas Front-end BR',
  remoteok: 'RemoteOK',
  remotive: 'Remotive',
}

export const MANUAL_SOURCE = 'manual'

/**
 * A readable name for a source slug.
 *
 * An unknown slug falls back to its own capitalised form rather than to a
 * placeholder: a board added to the worker keeps working in the interface
 * before anyone remembers to add it to the map above.
 */
export function sourceLabel(slug: string): string {
  return SOURCE_LABELS[slug] ?? slug.charAt(0).toUpperCase() + slug.slice(1)
}

/** Whether the opening was typed into the form rather than ingested. */
export function isManual(opening: Opening): boolean {
  return opening.source === MANUAL_SOURCE
}

/**
 * Whether the opening's content can be edited here.
 *
 * An ingested opening is a copy of a record the board owns: the next run would
 * overwrite any edit made to it. Deleting one is still allowed — that is how a
 * person dismisses it, and the upsert deliberately does not bring it back.
 */
export function isEditable(opening: Opening): boolean {
  return isManual(opening)
}
