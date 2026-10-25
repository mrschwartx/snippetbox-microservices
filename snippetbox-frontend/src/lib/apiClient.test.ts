import { afterEach, describe, expect, it, vi } from 'vitest'
import { createMemoryStorage } from '@/test/memoryStorage'
import { NETWORK_ERROR_MESSAGE, createApiClient } from './apiClient'
import {
  InvalidCredentialsError,
  NoRecordError,
  UnauthenticatedError,
  UnexpectedError,
  ValidationError,
} from './errors'
import { createSessionStore } from './sessionStore'

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const setup = (response: Response | (() => Promise<Response>), baseUrl = '/api/v1') => {
  const fetchImpl = vi.fn(async () =>
    typeof response === 'function' ? response() : response.clone(),
  )
  const storage = createMemoryStorage()
  const sessionStore = createSessionStore(() => storage)
  const client = createApiClient({ baseUrl, sessionStore, fetchImpl })
  const lastCall = () => {
    const [url, init] = fetchImpl.mock.calls.at(-1) as unknown as [string, RequestInit]
    return { url, init, headers: init.headers as Record<string, string> }
  }
  return { client, sessionStore, fetchImpl, lastCall }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('createApiClient requests', () => {
  it('prefixes the base URL and ignores trailing slashes on it', async () => {
    const { client, lastCall } = setup(json(200, {}), '/api/v1///')

    await client.get('/snippets')

    expect(lastCall().url).toBe('/api/v1/snippets')
  })

  it('sends GET requests with an Accept header and no body', async () => {
    const { client, lastCall } = setup(json(200, []))

    await client.get('/snippets')

    expect(lastCall().init.method).toBe('GET')
    expect(lastCall().init.body).toBeUndefined()
    expect(lastCall().headers).toEqual({ Accept: 'application/json' })
  })

  it('sends POST bodies as JSON', async () => {
    const { client, lastCall } = setup(json(200, {}))

    await client.post('/auth/login', { body: { email: 'a@b.co' } })

    expect(lastCall().init.method).toBe('POST')
    expect(lastCall().init.body).toBe('{"email":"a@b.co"}')
    expect(lastCall().headers['Content-Type']).toBe('application/json')
  })

  it('omits the body and content type for a POST without one', async () => {
    const { client, lastCall } = setup(json(200, {}))

    await client.post('/auth/logout')

    expect(lastCall().init.body).toBeUndefined()
    expect(lastCall().headers['Content-Type']).toBeUndefined()
  })

  it('sends the bearer token only when auth is requested', async () => {
    const { client, sessionStore, lastCall } = setup(json(200, {}))
    sessionStore.write({ token: 'abc', expiresAt: '2030-01-01T00:00:00Z' })

    await client.get('/public')
    expect(lastCall().headers.Authorization).toBeUndefined()

    await client.get('/private', { auth: true })
    expect(lastCall().headers.Authorization).toBe('Bearer abc')
  })

  it('sends no Authorization header when auth is requested but nobody is signed in', async () => {
    const { client, lastCall } = setup(json(200, {}))

    await client.get('/private', { auth: true })

    expect(lastCall().headers.Authorization).toBeUndefined()
  })

  it('uses the global fetch when none is injected', async () => {
    const globalFetch = vi.fn(async () => json(200, { ok: true }))
    vi.stubGlobal('fetch', globalFetch)
    const client = createApiClient({
      baseUrl: '/api/v1',
      sessionStore: createSessionStore(() => createMemoryStorage()),
    })

    await expect(client.get('/ping')).resolves.toEqual({ ok: true })
    expect(globalFetch).toHaveBeenCalledWith(
      '/api/v1/ping',
      expect.objectContaining({ method: 'GET' }),
    )
  })
})

describe('createApiClient responses', () => {
  it('returns the parsed JSON body', async () => {
    const { client } = setup(json(200, { id: 1 }))

    await expect(client.get('/snippets/1')).resolves.toEqual({ id: 1 })
  })

  it('returns undefined for 204 responses', async () => {
    const { client } = setup(new Response(null, { status: 204 }))

    await expect(client.post('/auth/logout')).resolves.toBeUndefined()
  })

  it('reports an unreadable success body as an unexpected error', async () => {
    const { client } = setup(new Response('<html>', { status: 200 }))

    await expect(client.get('/x')).rejects.toBeInstanceOf(UnexpectedError)
  })
})

describe('createApiClient error mapping', () => {
  const failWith = async (response: Response) => {
    const { client } = setup(response)
    return client.get('/x').catch((error: unknown) => error)
  }

  it('maps 401 with INVALID_CREDENTIALS to InvalidCredentialsError', async () => {
    const error = await failWith(json(401, { error: { code: 'INVALID_CREDENTIALS' } }))

    expect(error).toBeInstanceOf(InvalidCredentialsError)
  })

  it.each([
    ['another code', json(401, { error: { code: 'UNAUTHENTICATED' } })],
    ['an empty body', json(401, {})],
    ['a non-JSON body', new Response('nope', { status: 401 })],
  ])('maps 401 with %s to UnauthenticatedError', async (_label, response) => {
    expect(await failWith(response)).toBeInstanceOf(UnauthenticatedError)
  })

  it('maps 404 to NoRecordError', async () => {
    expect(await failWith(json(404, { error: { code: 'NOT_FOUND' } }))).toBeInstanceOf(
      NoRecordError,
    )
  })

  it('maps 422 to a ValidationError with the server field and non-field errors', async () => {
    const error = await failWith(
      json(422, { error: { fields: { title: 'bad' }, non_field: ['oops'] } }),
    )

    expect(error).toBeInstanceOf(ValidationError)
    expect((error as ValidationError).fieldErrors).toEqual({ title: 'bad' })
    expect((error as ValidationError).nonFieldErrors).toEqual(['oops'])
  })

  it('maps a 422 without details to an empty ValidationError', async () => {
    const error = (await failWith(json(422, {}))) as ValidationError

    expect(error).toBeInstanceOf(ValidationError)
    expect(error.fieldErrors).toEqual({})
    expect(error.nonFieldErrors).toEqual([])
  })

  it.each([
    ['a 500 error', json(500, { error: { code: 'INTERNAL_ERROR' } })],
    ['a 403 error', json(403, { error: { code: 'FORBIDDEN' } })],
    ['a proxy error page', new Response('<html>Bad gateway</html>', { status: 502 })],
  ])('maps %s to UnexpectedError', async (_label, response) => {
    expect(await failWith(response)).toBeInstanceOf(UnexpectedError)
  })

  it('reports network failures with a dedicated message', async () => {
    const { client } = setup(() => Promise.reject(new TypeError('offline')))

    const error = await client.get('/x').catch((caught: unknown) => caught)

    expect(error).toBeInstanceOf(UnexpectedError)
    expect((error as UnexpectedError).message).toBe(NETWORK_ERROR_MESSAGE)
  })
})
