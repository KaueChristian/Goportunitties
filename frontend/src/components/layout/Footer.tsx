import styles from './Footer.module.css'
import { Icon } from '../ui/Icon'
import { ROUTE_PATHS } from '../../router/useHashRoute'

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <div className={styles.about}>
          <a className={styles.brand} href={ROUTE_PATHS.home}>
            <span className={styles.mark}>
              <Icon name="compass" size={22} />
            </span>
            <span className={styles.wordmark}>
              Go<strong>portunitties</strong>
            </span>
          </a>
          <p className={styles.tagline}>
            Um indexador aberto de vagas de tecnologia. Publique, filtre e acompanhe as
            oportunidades que interessam sem depender de dez abas abertas.
          </p>
        </div>

        <nav className={styles.column} aria-label="Navegação do rodapé">
          <h4 className={styles.columnTitle}>Navegação</h4>
          <a href={ROUTE_PATHS.home}>Início</a>
          <a href={ROUTE_PATHS.jobs}>Vagas</a>
          <a href={ROUTE_PATHS.dashboard}>Painel</a>
        </nav>

        <div className={styles.column}>
          <h4 className={styles.columnTitle}>Projeto</h4>
          <a
            href="https://github.com/KaueChristian/Goportunitties"
            target="_blank"
            rel="noopener noreferrer"
          >
            Repositório
          </a>
          <a href="/api/v1/openings" target="_blank" rel="noopener noreferrer">
            API de vagas
          </a>
        </div>
      </div>

      <div className={styles.bottom}>
        <span>© {new Date().getFullYear()} Goportunitties</span>
        <span>Feito com Go e React</span>
      </div>
    </footer>
  )
}
