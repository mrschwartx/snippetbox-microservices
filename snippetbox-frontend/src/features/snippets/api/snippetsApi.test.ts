import { describe, expect, it } from 'vitest'
import { NoRecordError, UnauthenticatedError, ValidationError } from '@/lib/errors'
import { sessionStore } from '@/lib/sessionStore'
import { signIn } from '@/test/helpers'
import { backend } from '@/test/server'
import { advanceHours } from '@/test/time'
import { createSnippet, getLatestSnippets, getSnippet } from './snippetsApi'

const draft = { title: 'Hello', content: 'fmt.Println()', expires: 7 }

describe('getLatestSnippets', () => {
  it('returns snippets newest first', async () => {
    backend.seedSnippet({ title: 'older' })
    backend.seedSnippet({ title: 'newer' })

    const snippets = await getLatestSnippets()

    expect(snippets.map((snippet) => snippet.title)).toEqual(['newer', 'older'])
  })

  it('caps the list at ten and leaves out expired snippets', async () => {
    backend.seedSnippet({ title: 'short lived', days: 1 })
    for (let index = 0; index < 11; index++) backend.seedSnippet({ title: `snippet ${index}` })
    advanceHours(25)

    const snippets = await getLatestSnippets()

    expect(snippets).toHaveLength(10)
    expect(snippets[0].id).toBe(12)
    expect(snippets.some((snippet) => snippet.title === 'short lived')).toBe(false)
  })

  it('is empty when there are no snippets', async () => {
    await expect(getLatestSnippets()).resolves.toEqual([])
  })
})

describe('getSnippet', () => {
  it('returns the snippet with that id', async () => {
    backend.seedSnippet({ title: 'Greeting', content: 'hello' })

    await expect(getSnippet(1)).resolves.toMatchObject({ id: 1, title: 'Greeting' })
  })

  it.each([0, -1, 1.5, Number.NaN, Number.POSITIVE_INFINITY])(
    'treats %s as not found without calling the server',
    async (id) => {
      await expect(getSnippet(id)).rejects.toBeInstanceOf(NoRecordError)

      expect(backend.requests()).toEqual([])
    },
  )

  it('reports an unknown id as not found', async () => {
    await expect(getSnippet(99)).rejects.toBeInstanceOf(NoRecordError)
  })

  it('reports an expired snippet as not found', async () => {
    backend.seedSnippet({ title: 'short lived', days: 1 })
    advanceHours(25)

    await expect(getSnippet(1)).rejects.toBeInstanceOf(NoRecordError)
  })
})

describe('createSnippet', () => {
  it('refuses to create anything for an anonymous visitor, without calling the server', async () => {
    await expect(createSnippet(draft)).rejects.toBeInstanceOf(UnauthenticatedError)

    expect(backend.requests()).toEqual([])
  })

  it('creates a snippet that expires after the chosen number of days', async () => {
    await signIn()

    const snippet = await createSnippet(draft)

    expect(snippet).toMatchObject({
      id: 1,
      title: 'Hello',
      created: '2026-10-15T08:00:00.000Z',
      expires: '2026-10-22T08:00:00.000Z',
    })
  })

  it('rejects an invalid draft without calling the server', async () => {
    await signIn()
    backend.clearRequests()

    const error = await createSnippet({ title: '', content: '', expires: 3 }).catch(
      (e: unknown) => e,
    )

    expect(error).toBeInstanceOf(ValidationError)
    expect(Object.keys((error as ValidationError).fieldErrors)).toEqual([
      'title',
      'content',
      'expires',
    ])
    expect(backend.requests()).toEqual([])
  })

  it('reports a session the server does not recognise as unauthenticated', async () => {
    sessionStore.write({ token: 'unknown', expiresAt: '2030-01-01T00:00:00.000Z' })

    await expect(createSnippet(draft)).rejects.toBeInstanceOf(UnauthenticatedError)
  })
})
