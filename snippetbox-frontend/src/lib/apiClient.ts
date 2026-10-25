import {
  InvalidCredentialsError,
  NoRecordError,
  UnauthenticatedError,
  UnexpectedError,
  ValidationError,
} from './errors'
import { sessionStore as defaultSessionStore, type SessionStore } from './sessionStore'

export type FetchLike = (input: string, init?: RequestInit) => Promise<Response>

interface ErrorEnvelope {
  error?: {
    code?: string
    message?: string
    fields?: Record<string, string>
    non_field?: string[]
  }
}

interface RequestOptions {
  body?: unknown
  auth?: boolean
}

interface ApiClientConfig {
  baseUrl: string
  sessionStore: SessionStore
  fetchImpl?: FetchLike
}

export const NETWORK_ERROR_MESSAGE = 'Unable to reach the server. Please try again.'

const readErrorEnvelope = async (response: Response): Promise<ErrorEnvelope> => {
  try {
    return (await response.json()) as ErrorEnvelope
  } catch {
    return {}
  }
}

const toApiError = async (response: Response): Promise<Error> => {
  const { error } = await readErrorEnvelope(response)

  switch (response.status) {
    case 401:
      return error?.code === 'INVALID_CREDENTIALS'
        ? new InvalidCredentialsError()
        : new UnauthenticatedError()
    case 404:
      return new NoRecordError()
    case 422:
      return new ValidationError(error?.fields ?? {}, error?.non_field ?? [])
    default:
      return new UnexpectedError()
  }
}

export const createApiClient = ({
  baseUrl,
  sessionStore,
  fetchImpl = (input, init) => fetch(input, init),
}: ApiClientConfig) => {
  const root = baseUrl.replace(/\/+$/, '')

  const request = async <T>(
    method: string,
    path: string,
    { body, auth = false }: RequestOptions = {},
  ): Promise<T> => {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (body !== undefined) headers['Content-Type'] = 'application/json'

    const session = auth ? sessionStore.read() : null
    if (session) headers.Authorization = `Bearer ${session.token}`

    let response: Response
    try {
      response = await fetchImpl(`${root}${path}`, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      })
    } catch {
      throw new UnexpectedError(NETWORK_ERROR_MESSAGE)
    }

    if (!response.ok) throw await toApiError(response)
    if (response.status === 204) return undefined as T

    try {
      return (await response.json()) as T
    } catch {
      throw new UnexpectedError()
    }
  }

  return {
    get: <T>(path: string, options?: Omit<RequestOptions, 'body'>) =>
      request<T>('GET', path, options),
    post: <T>(path: string, options?: RequestOptions) => request<T>('POST', path, options),
  }
}

export type ApiClient = ReturnType<typeof createApiClient>

export const apiClient = createApiClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
  sessionStore: defaultSessionStore,
})
