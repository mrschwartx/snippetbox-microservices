import { vi } from 'vitest'

const HOUR_MS = 60 * 60 * 1000

export const TEST_NOW = new Date('2026-10-15T08:00:00.000Z')

export const freezeTime = (): void => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(TEST_NOW)
}

export const advanceHours = (hours: number): void => {
  vi.setSystemTime(new Date(Date.now() + hours * HOUR_MS))
}
