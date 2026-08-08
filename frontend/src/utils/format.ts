const currency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
  maximumFractionDigits: 0,
})

const compactCurrency = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
  notation: 'compact',
  maximumFractionDigits: 1,
})

const fullDate = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
})

const relative = new Intl.RelativeTimeFormat('pt-BR', { numeric: 'auto' })

export function formatSalary(value: number): string {
  return currency.format(value)
}

/** Compact form for stat tiles, where the digits are large and space is tight. */
export function formatSalaryCompact(value: number): string {
  return compactCurrency.format(value)
}

export function formatDate(iso: string): string {
  return fullDate.format(new Date(iso))
}

/** "há 3 dias" / "hoje" — the phrasing job boards use for posting age. */
export function formatRelativeDate(iso: string): string {
  const then = new Date(iso).getTime()
  const diffDays = Math.round((then - Date.now()) / 86_400_000)

  if (diffDays === 0) return 'hoje'
  if (diffDays > -30) return relative.format(diffDays, 'day')
  if (diffDays > -365) return relative.format(Math.round(diffDays / 30), 'month')
  return relative.format(Math.round(diffDays / 365), 'year')
}

/** Up to two initials, used for the company avatar. */
export function initialsOf(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) return '?'
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase()
  return (words[0][0] + words[1][0]).toUpperCase()
}

/**
 * Deterministic avatar hue per company: the same name always gets the same
 * color, and the set is a fixed ordered palette rather than a generated hue.
 */
const AVATAR_HUES = [
  'var(--blue-450)',
  '#1baf7a',
  '#7c5cd6',
  '#eb6834',
  '#e87ba4',
  '#0f8ea3',
] as const

export function avatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i += 1) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  }
  return AVATAR_HUES[hash % AVATAR_HUES.length]
}

/** Normalizes for accent- and case-insensitive search. */
export function normalize(value: string): string {
  return value
    .toLowerCase()
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
}
