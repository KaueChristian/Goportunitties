import { afterEach, describe, expect, it, vi } from 'vitest'
import { getAdminKey } from './adminKey'

afterEach(() => {
  localStorage.clear()
})

describe('getAdminKey', () => {
  it('returns an empty string when nothing is stored', () => {
    expect(getAdminKey()).toBe('')
  })

  it('returns what was stored', () => {
    localStorage.setItem('goportunitties:admin-key', 's3cr3t')

    expect(getAdminKey()).toBe('s3cr3t')
  })

  // A locked-down browsing context (strict private mode, a sandboxed iframe)
  // can make localStorage throw on access rather than returning null.
  it('degrades to an empty key instead of throwing', () => {
    const spy = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage disabled')
    })

    expect(getAdminKey()).toBe('')

    spy.mockRestore()
  })
})
