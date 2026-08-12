import { useCallback, useEffect, useState } from 'react'

export type Route = 'home' | 'jobs' | 'dashboard'

/** The hash each route owns. Anything unrecognized falls back to home. */
export const ROUTE_PATHS: Record<Route, string> = {
  home: '#/',
  jobs: '#/vagas',
  dashboard: '#/painel',
}

const BY_PATH = new Map<string, Route>(
  Object.entries(ROUTE_PATHS).map(([route, path]) => [path, route as Route]),
)

function readRoute(): Route {
  return BY_PATH.get(window.location.hash) ?? 'home'
}

/**
 * Hash routing on purpose: the built bundle is served as static files by the Go
 * binary, which has no catch-all rewrite to hand deep paths back to index.html.
 * Hashes never reach the server, so every route survives a reload.
 */
export function useHashRoute() {
  const [route, setRoute] = useState<Route>(readRoute)

  useEffect(() => {
    // Also fires for programmatic changes, so back/forward and nav clicks share
    // the same scroll-reset path.
    const sync = () => {
      setRoute(readRoute())
      window.scrollTo(0, 0)
    }

    window.addEventListener('hashchange', sync)
    return () => window.removeEventListener('hashchange', sync)
  }, [])

  const navigate = useCallback((next: Route) => {
    window.location.hash = ROUTE_PATHS[next]
  }, [])

  return { route, navigate }
}
