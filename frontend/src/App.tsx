import styles from './App.module.css'

import { Header } from './components/layout/Header'
import { Footer } from './components/layout/Footer'
import { HomePage } from './pages/HomePage'
import { JobsPage } from './pages/JobsPage'
import { DashboardPage } from './pages/DashboardPage'

import { useTheme } from './hooks/useTheme'
import { useHashRoute } from './router/useHashRoute'
import { useDialogs } from './state/DialogsProvider'

/**
 * The application shell: theme, route and layout.
 *
 * Everything about the openings — filters, the page on screen, the writes, the
 * modals — lives in the providers around this component, so a page reads what
 * it needs through a hook instead of receiving it as a prop.
 */
export default function App() {
  const { theme, toggleTheme } = useTheme()
  const { route } = useHashRoute()
  const { openCreate } = useDialogs()

  return (
    <div className={styles.app}>
      <Header route={route} theme={theme} onToggleTheme={toggleTheme} onCreate={openCreate} />

      <main className={styles.main}>
        {route === 'home' && <HomePage />}
        {route === 'jobs' && <JobsPage />}
        {route === 'dashboard' && <DashboardPage />}
      </main>

      <Footer />
    </div>
  )
}
