import { describe, expect, it } from 'vitest'
import { createBlockedStorage, createMemoryStorage } from '@/test/memoryStorage'
import {
  SESSION_EXPIRY_SKEW_MS,
  SESSION_STORAGE_KEY,
  createSessionStore,
  isSessionActive,
  sessionStore,
} from './sessionStore'

const session = { token: 'abc', expiresAt: '2030-01-01T00:00:00.000Z' }

describe('createSessionStore', () => {
  it('has nothing stored initially', () => {
    expect(createSessionStore(() => createMemoryStorage()).read()).toBeNull()
  })

  it('round-trips a session', () => {
    const storage = createMemoryStorage()
    const store = createSessionStore(() => storage)

    store.write(session)

    expect(store.read()).toEqual(session)
  })

  it('stores the session as JSON under a namespaced key', () => {
    const storage = createMemoryStorage()

    createSessionStore(() => storage).write(session)

    expect(JSON.parse(storage.data.get(SESSION_STORAGE_KEY) ?? '')).toEqual(session)
  })

  it('forgets the session once cleared', () => {
    const storage = createMemoryStorage()
    const store = createSessionStore(() => storage)
    store.write(session)

    store.clear()

    expect(store.read()).toBeNull()
    expect(storage.data.size).toBe(0)
  })

  it.each([
    ['broken JSON', '{not json'],
    ['a JSON null', 'null'],
    ['a JSON string', '"token"'],
    ['a numeric token', JSON.stringify({ token: 1, expiresAt: 'x' })],
    ['a missing expiry', JSON.stringify({ token: 'abc' })],
    ['an empty string', ''],
  ])('treats %s as no session', (_label, raw) => {
    const storage = createMemoryStorage()
    storage.data.set(SESSION_STORAGE_KEY, raw)

    expect(createSessionStore(() => storage).read()).toBeNull()
  })

  it('never throws when the storage methods are blocked', () => {
    const store = createSessionStore(() => createBlockedStorage())

    expect(store.read()).toBeNull()
    expect(() => store.write(session)).not.toThrow()
    expect(() => store.clear()).not.toThrow()
  })

  it('never throws when the storage itself cannot be reached', () => {
    const store = createSessionStore(() => {
      throw new DOMException('denied', 'SecurityError')
    })

    expect(store.read()).toBeNull()
    expect(() => store.write(session)).not.toThrow()
    expect(() => store.clear()).not.toThrow()
  })
})

describe('the default sessionStore', () => {
  it('persists to window.localStorage', () => {
    sessionStore.write(session)

    expect(JSON.parse(window.localStorage.getItem(SESSION_STORAGE_KEY) ?? '')).toEqual(session)
    expect(sessionStore.read()).toEqual(session)
  })
})

describe('isSessionActive', () => {
  const now = new Date('2026-10-15T08:00:00.000Z')
  const expiringIn = (ms: number) => ({
    token: 't',
    expiresAt: new Date(now.getTime() + ms).toISOString(),
  })

  it('is active well before expiry', () => {
    expect(isSessionActive(expiringIn(60 * 60 * 1000), now)).toBe(true)
  })

  it('is inactive once the expiry has passed', () => {
    expect(isSessionActive(expiringIn(-1000), now)).toBe(false)
  })

  it('stops being active five seconds early to absorb clock drift', () => {
    expect(isSessionActive(expiringIn(5_001), now)).toBe(true)
    expect(isSessionActive(expiringIn(5_000), now)).toBe(false)
    expect(isSessionActive(expiringIn(4_999), now)).toBe(false)
  })

  it('exposes the skew it applies', () => {
    expect(SESSION_EXPIRY_SKEW_MS).toBe(5_000)
  })

  it('is inactive when the expiry is not a valid date', () => {
    expect(isSessionActive({ token: 't', expiresAt: 'garbage' }, now)).toBe(false)
  })

  it('compares against the current time by default', () => {
    expect(isSessionActive(expiringIn(60 * 60 * 1000))).toBe(true)
    expect(isSessionActive(expiringIn(-60 * 60 * 1000))).toBe(false)
  })
})
