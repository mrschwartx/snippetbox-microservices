import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll, beforeEach, vi } from 'vitest'
import { backend, server } from './server'
import { freezeTime } from './time'

beforeAll(() => server.listen({ onUnhandledFrame: 'error' }))

beforeEach(() => {
  freezeTime()
})

afterEach(() => {
  cleanup()
  server.resetHandlers()
  backend.reset()
  window.localStorage.clear()
  vi.useRealTimers()
})

afterAll(() => server.close())
