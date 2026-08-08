/**
 * Inline icon set. Keeping the paths here (instead of an icon dependency) means
 * one file to audit, no runtime fetch, and every icon inherits currentColor.
 */

export type IconName =
  | 'briefcase'
  | 'search'
  | 'plus'
  | 'edit'
  | 'trash'
  | 'close'
  | 'external'
  | 'pin'
  | 'building'
  | 'remote'
  | 'sun'
  | 'moon'
  | 'check'
  | 'alert'
  | 'inbox'
  | 'salary'
  | 'clock'
  | 'chevron-down'
  | 'arrow-right'
  | 'grid'
  | 'list'
  | 'bookmark'
  | 'users'
  | 'chart'
  | 'home'
  | 'filter'
  | 'send'
  | 'logout'
  | 'trending'
  | 'menu'
  | 'eye'
  | 'radar'
  | 'compass'
  | 'compass-rose'
  | 'scope'

/** Exported so the favicon can be rebuilt from the same path the header draws. */
export const ICON_PATHS: Record<IconName, string> = {
  briefcase:
    'M3 8.5A2.5 2.5 0 0 1 5.5 6h13A2.5 2.5 0 0 1 21 8.5v9a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 17.5v-9Zm6-2V5.5A1.5 1.5 0 0 1 10.5 4h3A1.5 1.5 0 0 1 15 5.5V6M3 12h18',
  search: 'M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14Zm5 12 4 4',
  plus: 'M12 5v14M5 12h14',
  edit: 'M4 20h4L18.5 9.5a2.1 2.1 0 0 0-3-3L5 17v3Zm10-12 3 3',
  trash: 'M4 7h16M9 7V5.5A1.5 1.5 0 0 1 10.5 4h3A1.5 1.5 0 0 1 15 5.5V7m3 0v12a1.5 1.5 0 0 1-1.5 1.5h-9A1.5 1.5 0 0 1 6 19V7m4 4v6m4-6v6',
  close: 'M6 6 18 18M18 6 6 18',
  external: 'M14 5h5v5M19 5l-8 8M18 14v4.5a1.5 1.5 0 0 1-1.5 1.5h-11A1.5 1.5 0 0 1 4 18.5v-11A1.5 1.5 0 0 1 5.5 6H10',
  pin: 'M12 21s7-5.5 7-11a7 7 0 1 0-14 0c0 5.5 7 11 7 11Zm0-8.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
  building: 'M4 20h16M6 20V5.5A1.5 1.5 0 0 1 7.5 4h6A1.5 1.5 0 0 1 15 5.5V20M15 10h2.5A1.5 1.5 0 0 1 19 11.5V20M9 8h3M9 12h3M9 16h3',
  remote: 'M5 12.5a9.5 9.5 0 0 1 14 0M8 16a5.5 5.5 0 0 1 8 0M12 19.5h.01M2 9a14 14 0 0 1 20 0',
  sun: 'M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Zm0-5v2m0 14v2M3 12h2m14 0h2M5.6 5.6l1.4 1.4m10 10 1.4 1.4m0-12.8-1.4 1.4m-10 10-1.4 1.4',
  moon: 'M20 14.5A8.5 8.5 0 0 1 9.5 4a8.5 8.5 0 1 0 10.5 10.5Z',
  check: 'm5 13 4 4L19 7',
  alert: 'M12 8v5m0 3h.01M10.3 4.3 2.5 18a1.5 1.5 0 0 0 1.3 2.2h16.4a1.5 1.5 0 0 0 1.3-2.2L13.7 4.3a2 2 0 0 0-3.4 0Z',
  inbox:
    'M4 13h4l1.5 3h5L16 13h4M4 13l2.3-7A1.5 1.5 0 0 1 7.7 5h8.6a1.5 1.5 0 0 1 1.4 1L20 13v5a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 18v-5Z',
  salary:
    'M12 3v18M16 7.5C16 6 14.2 5 12 5S8 6 8 7.5 9.8 10 12 10.5s4 1 4 2.5-1.8 2.5-4 2.5-4-1-4-2.5',
  clock: 'M12 4a8 8 0 1 0 0 16 8 8 0 0 0 0-16Zm0 3.5V12l3 2',
  'chevron-down': 'm6 9.5 6 6 6-6',
  'arrow-right': 'M4 12h15m-6-6 6 6-6 6',
  grid:
    'M4 5.5A1.5 1.5 0 0 1 5.5 4h4A1.5 1.5 0 0 1 11 5.5v4A1.5 1.5 0 0 1 9.5 11h-4A1.5 1.5 0 0 1 4 9.5v-4Zm9 0A1.5 1.5 0 0 1 14.5 4h4A1.5 1.5 0 0 1 20 5.5v4a1.5 1.5 0 0 1-1.5 1.5h-4A1.5 1.5 0 0 1 13 9.5v-4Zm-9 9A1.5 1.5 0 0 1 5.5 13h4a1.5 1.5 0 0 1 1.5 1.5v4A1.5 1.5 0 0 1 9.5 20h-4A1.5 1.5 0 0 1 4 18.5v-4Zm9 0a1.5 1.5 0 0 1 1.5-1.5h4a1.5 1.5 0 0 1 1.5 1.5v4a1.5 1.5 0 0 1-1.5 1.5h-4a1.5 1.5 0 0 1-1.5-1.5v-4Z',
  list: 'M9 6h11M9 12h11M9 18h11M4.5 6h.01M4.5 12h.01M4.5 18h.01',
  bookmark: 'M7 4.5h10a1.5 1.5 0 0 1 1.5 1.5v14l-6.5-4-6.5 4V6A1.5 1.5 0 0 1 7 4.5Z',
  users:
    'M16 20v-1.5a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4V20M9.5 10.5a3.25 3.25 0 1 0 0-6.5 3.25 3.25 0 0 0 0 6.5ZM21 20v-1.5a4 4 0 0 0-3-3.87M16.5 4.13a4 4 0 0 1 0 7.75',
  chart: 'M4 19.5h16M7.5 16.5V11m4.5 5.5V5.5m4.5 11V13',
  home: 'M4 10.5 12 4l8 6.5V19a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 19v-8.5Zm5.5 10v-6h5v6',
  filter: 'M4 8h10m4 0h2M4 16h4m4 0h8M16 5.5v5M8 13.5v5',
  send: 'M21 3 10.5 13.5M21 3l-6.5 18-4-8-8-4L21 3Z',
  logout:
    'M15 16.5V19a1.5 1.5 0 0 1-1.5 1.5h-8A1.5 1.5 0 0 1 4 19V5a1.5 1.5 0 0 1 1.5-1.5h8A1.5 1.5 0 0 1 15 5v2.5M10 12h11m0 0-3.5-3.5M21 12l-3.5 3.5',
  trending: 'M4 16.5 9.5 11l3.5 3.5L20 7m0 0h-5.5M20 7v5.5',
  menu: 'M4 7h16M4 12h16M4 17h16',
  eye:
    'M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Zm9.5 2.75a2.75 2.75 0 1 0 0-5.5 2.75 2.75 0 0 0 0 5.5Z',

  /* ---- Brand mark candidates ---- */
  // Scope with a sweep arm ending exactly on the rim, plus a contact blip.
  radar:
    'M12 3.5a8.5 8.5 0 1 0 0 17 8.5 8.5 0 0 0 0-17ZM12 8.5a3.5 3.5 0 1 0 0 7 3.5 3.5 0 0 0 0-7ZM12 12 18 6M15.8 15.8h.01',
  // Needle whose two tips sit 6.5 from the centre, so it reads level in the dial.
  compass: 'M12 3.5a8.5 8.5 0 1 0 0 17 8.5 8.5 0 0 0 0-17ZM16.6 7.4l-2.4 7.2-7.2 2.4 2.4-7.2 7.2-2.4Z',
  'compass-rose':
    'M12 3.5a8.5 8.5 0 1 0 0 17 8.5 8.5 0 0 0 0-17ZM12 6l1.7 4.3 4.3 1.7-4.3 1.7-1.7 4.3-1.7-4.3-4.3-1.7 4.3-1.7 1.7-4.3Z',
  scope:
    'M12 4.5a7.5 7.5 0 1 0 0 15 7.5 7.5 0 0 0 0-15ZM12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6ZM12 2v3M12 19v3M2 12h3M19 12h3',
}

interface IconProps {
  name: IconName
  size?: number
  className?: string
}

export function Icon({ name, size = 18, className }: IconProps) {
  return (
    <svg
      className={className}
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d={ICON_PATHS[name]} />
    </svg>
  )
}
