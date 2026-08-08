import { useCallback, useEffect, useState } from 'react'
import { ICON_PATHS, type IconName } from '../ui/Icon'

/*
 * ANDAIME TEMPORÁRIO — existe só para comparar as marcas candidatas no header
 * real. Quando a escolha sair, este arquivo some: o header e o rodapé passam a
 * usar o ícone escolhido direto, e o favicon volta a ser o SVG estático.
 */

export const BRAND_MARKS: { icon: IconName; label: string }[] = [
  { icon: 'radar', label: 'Radar' },
  { icon: 'compass', label: 'Bússola' },
  { icon: 'compass-rose', label: 'Rosa dos ventos' },
  { icon: 'scope', label: 'Mira' },
]

const STORAGE_KEY = 'goportunitties:brand-mark'
const CHANGE_EVENT = 'goportunitties:brand-mark-change'

const ACCENT = '#2a78d6'

function readIndex(): number {
  const stored = Number(localStorage.getItem(STORAGE_KEY))
  return Number.isInteger(stored) && stored >= 0 && stored < BRAND_MARKS.length ? stored : 0
}

/**
 * Redesenha o favicon com a marca atual. O glifo é reduzido a 60% e centrado,
 * senão ele encosta na borda do quadrado; o traço é engrossado antes da escala
 * para não sumir nos 16px da aba.
 */
function paintFavicon(icon: IconName) {
  const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) return

  const svg = [
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">',
    `<rect width="24" height="24" rx="6" fill="${ACCENT}"/>`,
    '<g transform="translate(4.8 4.8) scale(0.6)">',
    `<path d="${ICON_PATHS[icon]}" fill="none" stroke="#fff" stroke-width="2.8"`,
    ' stroke-linecap="round" stroke-linejoin="round"/>',
    '</g></svg>',
  ].join('')

  link.href = `data:image/svg+xml,${encodeURIComponent(svg)}`
}

export function useBrandMark() {
  const [index, setIndex] = useState(readIndex)

  // Header e rodapé mantêm cópias do estado; o evento mantém as duas iguais.
  useEffect(() => {
    const sync = () => setIndex(readIndex())
    window.addEventListener(CHANGE_EVENT, sync)
    return () => window.removeEventListener(CHANGE_EVENT, sync)
  }, [])

  useEffect(() => paintFavicon(BRAND_MARKS[index].icon), [index])

  const cycle = useCallback(() => {
    const next = (readIndex() + 1) % BRAND_MARKS.length
    localStorage.setItem(STORAGE_KEY, String(next))
    window.dispatchEvent(new Event(CHANGE_EVENT))
  }, [])

  return { mark: BRAND_MARKS[index], index, cycle }
}
