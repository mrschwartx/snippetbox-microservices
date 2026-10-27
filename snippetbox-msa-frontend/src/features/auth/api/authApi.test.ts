import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { InvalidCredentialsError, UnexpectedError, ValidationError } from '@/lib/errors'
import { sessionStore } from '@/lib/sessionStore'
import { DEFAULT_USER } from '@/test/fakeBackend'
import { signIn } from '@/test/helpers'
import { apiUrl, backend, server } from '@/test/server'
import { advanceHours } from '@/test/time'
import { getCurrentUser, login, logout, signup } from './authApi'

describe('signup', () => {
  it('creates the account and returns the user', async () => {
    const user = await signup({ ...DEFAULT_USER, name: ' Alice ', email: ' alice@example.com ' })

    expect(user).toMatchObject({ id: 1, name: 'Alice', email: 'alice@example.com' })
    expect(backend.requests()).toEqual(['POST /auth/signup'])
  })

  it('does not sign the new user in', async () => {
    await signup(DEFAULT_USER)

    expect(sessionStore.read()).toBeNull()
  })

  it('rejects invalid input without calling the server', async () => {
    const error = await signup({ name: '', email: '', password: '' }).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ValidationError)
    expect(Object.keys((error as ValidationError).fieldErrors)).toEqual([
      'name',
      'email',
      'password',
    ])
    expect(backend.requests()).toEqual([])
  })

  it('surfaces a duplicate email reported by the server on the email field', async () => {
    backend.seedUser()

    const error = await signup({ ...DEFAULT_USER, name: 'Eve' }).catch((e: unknown) => e)

    expect((error as ValidationError).fieldErrors).toEqual({
      email: 'Email address is already in use',
    })
  })
})

describe('login', () => {
  it('stores the session and returns the user', async () => {
    backend.seedUser()

    const user = await login({ email: DEFAULT_USER.email, password: DEFAULT_USER.password })

    expect(user.email).toBe(DEFAULT_USER.email)
    expect(sessionStore.read()).toEqual({
      token: expect.stringMatching(/^token-/),
      expiresAt: '2026-10-15T20:00:00.000Z',
    })
  })

  it('trims the email before sending it', async () => {
    backend.seedUser()

    await expect(
      login({ email: `  ${DEFAULT_USER.email}  `, password: DEFAULT_USER.password }),
    ).resolves.toMatchObject({ email: DEFAULT_USER.email })
    expect(sessionStore.read()).not.toBeNull()
  })

  it('rejects a wrong password and stores nothing', async () => {
    backend.seedUser()

    const error = await login({ email: DEFAULT_USER.email, password: 'wrong-password' }).catch(
      (e: unknown) => e,
    )

    expect(error).toBeInstanceOf(InvalidCredentialsError)
    expect(sessionStore.read()).toBeNull()
  })

  it('rejects invalid input without calling the server', async () => {
    const error = await login({ email: '', password: '' }).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ValidationError)
    expect(backend.requests()).toEqual([])
  })
})

describe('logout', () => {
  it('tells the server and clears the session', async () => {
    await signIn()
    backend.clearRequests()

    await logout()

    expect(backend.requests()).toEqual(['POST /auth/logout'])
    expect(sessionStore.read()).toBeNull()
  })

  it('clears the session even when the server rejects the request', async () => {
    await signIn()
    server.use(http.post(apiUrl('/auth/logout'), () => HttpResponse.json({}, { status: 500 })))

    await expect(logout()).resolves.toBeUndefined()

    expect(sessionStore.read()).toBeNull()
  })

  it('clears the session even when the network is down', async () => {
    await signIn()
    server.use(http.post(apiUrl('/auth/logout'), () => HttpResponse.error()))

    await expect(logout()).resolves.toBeUndefined()

    expect(sessionStore.read()).toBeNull()
  })
})

describe('getCurrentUser', () => {
  it('is null without a session and does not call the server', async () => {
    await expect(getCurrentUser()).resolves.toBeNull()

    expect(backend.requests()).toEqual([])
  })

  it('returns the user for an active session', async () => {
    await signIn()

    await expect(getCurrentUser()).resolves.toMatchObject({ email: DEFAULT_USER.email })
  })

  it('drops an expired session without calling the server', async () => {
    await signIn()
    backend.clearRequests()

    advanceHours(13)

    await expect(getCurrentUser()).resolves.toBeNull()
    expect(sessionStore.read()).toBeNull()
    expect(backend.requests()).toEqual([])
  })

  it('drops a session the server no longer accepts', async () => {
    sessionStore.write({ token: 'revoked', expiresAt: '2030-01-01T00:00:00.000Z' })

    await expect(getCurrentUser()).resolves.toBeNull()
    expect(sessionStore.read()).toBeNull()
  })

  it('keeps the session and rethrows when the server fails', async () => {
    await signIn()
    server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json({}, { status: 500 })))

    await expect(getCurrentUser()).rejects.toBeInstanceOf(UnexpectedError)
    expect(sessionStore.read()).not.toBeNull()
  })

  it('keeps the session and rethrows when the network is down', async () => {
    await signIn()
    server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.error()))

    await expect(getCurrentUser()).rejects.toBeInstanceOf(UnexpectedError)
    expect(sessionStore.read()).not.toBeNull()
  })
})
