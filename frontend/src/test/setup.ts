import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'

// jsdom implements no layout, so it ships no scrollIntoView. Components that
// keep a highlighted option in view call it legitimately — stubbing the gap is
// right, weakening the component to accommodate the test environment is not.
if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = vi.fn()
}

// Testing Library does not unmount between tests on its own outside of its own
// globals setup, and a leftover tree would leak into the next query.
afterEach(() => {
  cleanup()
})
