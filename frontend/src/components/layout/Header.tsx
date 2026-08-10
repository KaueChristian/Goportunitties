import { useEffect, useState } from 'react'
import styles from './Header.module.css'
import { Button } from '../ui/Button'
import { Icon, type IconName } from '../ui/Icon'
import { ROUTE_PATHS, type Route } from '../../router/useHashRoute'
import type { Theme } from '../../hooks/useTheme'

interface HeaderProps {
  route: Route
  theme: Theme
  onToggleTheme: () => void
  onCreate: () => void
}

const NAV: { route: Route; label: string; icon: IconName }[] = [
  { route: 'home', label: 'Início', icon: 'home' },
  { route: 'jobs', label: 'Vagas', icon: 'briefcase' },
  { route: 'dashboard', label: 'Painel', icon: 'chart' },
]

export function Header({ route, theme, onToggleTheme, onCreate }: HeaderProps) {
  const [menuOpen, setMenuOpen] = useState(false)

  // The mobile sheet has no reason to outlive the navigation that was its point.
  useEffect(() => setMenuOpen(false), [route])

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <a className={styles.brand} href={ROUTE_PATHS.home}>
          <span className={styles.mark}>
            <Icon name="compass" size={22} />
          </span>
          <span className={styles.wordmark}>
            Go<strong>portunitties</strong>
          </span>
        </a>

        <nav className={styles.nav} aria-label="Navegação principal">
          {NAV.map((item) => (
            <a
              key={item.route}
              className={`${styles.link} ${route === item.route ? styles.linkActive : ''}`}
              href={ROUTE_PATHS[item.route]}
              aria-current={route === item.route ? 'page' : undefined}
            >
              {item.label}
            </a>
          ))}
        </nav>

        <div className={styles.actions}>
          <button
            className={styles.iconButton}
            onClick={onToggleTheme}
            aria-label={theme === 'dark' ? 'Ativar tema claro' : 'Ativar tema escuro'}
            title={theme === 'dark' ? 'Tema claro' : 'Tema escuro'}
          >
            <Icon name={theme === 'dark' ? 'sun' : 'moon'} size={18} />
          </button>

          <Button variant="primary" icon={<Icon name="plus" size={17} />} onClick={onCreate}>
            <span className={styles.ctaLabel}>Publicar vaga</span>
          </Button>

          <button
            className={`${styles.iconButton} ${styles.menuButton}`}
            onClick={() => setMenuOpen((open) => !open)}
            aria-label="Abrir menu de navegação"
            aria-expanded={menuOpen}
          >
            <Icon name={menuOpen ? 'close' : 'menu'} size={20} />
          </button>
        </div>
      </div>

      {menuOpen && (
        <nav className={styles.sheet} aria-label="Navegação principal (móvel)">
          {NAV.map((item) => (
            <a
              key={item.route}
              className={`${styles.sheetLink} ${route === item.route ? styles.sheetLinkActive : ''}`}
              href={ROUTE_PATHS[item.route]}
              aria-current={route === item.route ? 'page' : undefined}
            >
              <Icon name={item.icon} size={18} />
              {item.label}
            </a>
          ))}
        </nav>
      )}
    </header>
  )
}
