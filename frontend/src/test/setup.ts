import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Testing Library does not unmount between tests on its own outside of its own
// globals setup, and a leftover tree would leak into the next query.
afterEach(() => {
  cleanup()
})
