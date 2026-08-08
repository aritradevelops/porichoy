import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Not using Vitest's global test APIs (test.globals is off), so RTL's own
// auto-cleanup — which relies on detecting a global afterEach — doesn't run
// on its own. Wiring it explicitly instead.
afterEach(() => {
  cleanup()
})
