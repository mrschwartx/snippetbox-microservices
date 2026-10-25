import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { createFakeBackend } from './fakeBackend'

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string

const API_PATH_PREFIX = new URL(API_BASE_URL).pathname

export const backend = createFakeBackend()

const readBody = async (request: Request): Promise<unknown> => {
  const text = await request.text()
  return text ? JSON.parse(text) : undefined
}

const readToken = (request: Request): string | null =>
  request.headers.get('Authorization')?.replace(/^Bearer /, '') ?? null

const fakeApi = http.all(`${API_BASE_URL}/*`, async ({ request }) => {
  const { status, body } = backend.handle({
    method: request.method,
    path: new URL(request.url).pathname.slice(API_PATH_PREFIX.length),
    token: readToken(request),
    body: await readBody(request),
  })

  return body === undefined
    ? new HttpResponse(null, { status })
    : HttpResponse.json(body, { status })
})

export const server = setupServer(fakeApi)

export const apiUrl = (path: string): string => `${API_BASE_URL}${path}`
