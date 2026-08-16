const STORAGE_KEY = 'goportunitties:admin-key'

/**
 * The shared secret that unlocks create, edit and delete on the API — see
 * middleware.RequireAdminKey on the Go side.
 *
 * There is no UI to enter it: this project has one operator (whoever deploys
 * it), not a login system with users to manage. Setting it is one line in the
 * browser console:
 *
 *   localStorage.setItem('goportunitties:admin-key', 'a-chave-configurada-no-servidor')
 *
 * A visitor with no key keeps browsing normally — reading was never gated.
 */
export function getAdminKey(): string {
  try {
    return localStorage.getItem(STORAGE_KEY) ?? ''
  } catch {
    // Storage can throw in a locked-down browsing context (private mode with
    // strict settings, an iframe with storage disabled); an empty key just
    // means writes fall back to whatever the server does with none.
    return ''
  }
}
