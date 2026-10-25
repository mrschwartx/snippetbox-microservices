export interface FakeRequest {
  method: string
  path: string
  token?: string | null
  body?: unknown
}

export interface FakeResponse {
  status: number
  body?: unknown
}

export interface TestUser {
  name: string
  email: string
  password: string
}

export interface FakeBackendOptions {
  now?: () => Date
}

interface StoredUser extends TestUser {
  id: number
  created: string
}

interface StoredSnippet {
  id: number
  title: string
  content: string
  created: string
  expires: string
}

interface SnippetSeed {
  title: string
  content?: string
  days?: number
}

const HOUR_MS = 60 * 60 * 1000
const DAY_MS = 24 * HOUR_MS
const TOKEN_LIFETIME_MS = 12 * HOUR_MS
const LATEST_SNIPPETS_LIMIT = 10

export const DEFAULT_USER: TestUser = {
  name: 'Alice',
  email: 'alice@example.com',
  password: 'correct-horse',
}

const normalizeEmail = (email: unknown): string => String(email ?? '').toLowerCase()

export const createFakeBackend = ({ now = () => new Date() }: FakeBackendOptions = {}) => {
  let users: StoredUser[] = []
  let snippets: StoredSnippet[] = []
  let tokens = new Map<string, { userId: number; expiresAt: number }>()
  let requests: string[] = []

  const respond = (status: number, body?: unknown): FakeResponse => ({ status, body })

  const failure = (status: number, code: string, details: object = {}): FakeResponse =>
    respond(status, { error: { code, message: code, ...details } })

  const toPublicUser = ({ id, name, email, created }: StoredUser) => ({ id, name, email, created })

  const authenticate = (token?: string | null): StoredUser | null => {
    const record = token ? tokens.get(token) : undefined
    if (!record || record.expiresAt <= now().getTime()) return null
    return users.find((user) => user.id === record.userId) ?? null
  }

  const isLive = (snippet: StoredSnippet): boolean =>
    new Date(snippet.expires).getTime() > now().getTime()

  const addUser = ({ name, email, password }: TestUser): StoredUser => {
    const user: StoredUser = {
      id: users.length + 1,
      name,
      email: normalizeEmail(email),
      password,
      created: now().toISOString(),
    }
    users.push(user)
    return user
  }

  const addSnippet = ({ title, content = 'content', days = 7 }: SnippetSeed): StoredSnippet => {
    const created = now()
    const snippet: StoredSnippet = {
      id: snippets.length + 1,
      title,
      content,
      created: created.toISOString(),
      expires: new Date(created.getTime() + days * DAY_MS).toISOString(),
    }
    snippets.push(snippet)
    return snippet
  }

  const signup = (input: Record<string, unknown>): FakeResponse => {
    const email = normalizeEmail(input.email)
    if (users.some((user) => user.email === email)) {
      return failure(422, 'VALIDATION_ERROR', {
        fields: { email: 'Email address is already in use' },
      })
    }
    const user = addUser({
      name: String(input.name),
      email,
      password: String(input.password),
    })
    return respond(201, toPublicUser(user))
  }

  const login = (input: Record<string, unknown>): FakeResponse => {
    const user = users.find((candidate) => candidate.email === normalizeEmail(input.email))
    if (!user || user.password !== input.password) return failure(401, 'INVALID_CREDENTIALS')

    const expiresAt = now().getTime() + TOKEN_LIFETIME_MS
    const token = `token-${user.id}-${tokens.size + 1}`
    tokens.set(token, { userId: user.id, expiresAt })
    return respond(200, {
      token,
      token_type: 'Bearer',
      expires_at: new Date(expiresAt).toISOString(),
      user: toPublicUser(user),
    })
  }

  const createSnippet = (input: Record<string, unknown>): FakeResponse =>
    respond(
      201,
      addSnippet({
        title: String(input.title),
        content: String(input.content),
        days: Number(input.expires),
      }),
    )

  const handle = ({ method, path, token, body }: FakeRequest): FakeResponse => {
    requests.push(`${method} ${path}`)

    const input = (body ?? {}) as Record<string, unknown>
    const user = authenticate(token)
    const snippetId = /^\/snippets\/(\d+)$/.exec(path)?.[1]
    const route = `${method} ${snippetId ? '/snippets/:id' : path}`

    switch (route) {
      case 'POST /auth/signup':
        return signup(input)
      case 'POST /auth/login':
        return login(input)
      case 'POST /auth/logout':
        return user ? respond(204) : failure(401, 'UNAUTHENTICATED')
      case 'GET /auth/me':
        return user ? respond(200, toPublicUser(user)) : failure(401, 'UNAUTHENTICATED')
      case 'GET /snippets':
        return respond(
          200,
          snippets
            .filter(isLive)
            .sort((a, b) => b.id - a.id)
            .slice(0, LATEST_SNIPPETS_LIMIT),
        )
      case 'GET /snippets/:id': {
        const snippet = snippets.find((candidate) => candidate.id === Number(snippetId))
        return snippet && isLive(snippet) ? respond(200, snippet) : failure(404, 'NOT_FOUND')
      }
      case 'POST /snippets':
        return user ? createSnippet(input) : failure(401, 'UNAUTHENTICATED')
      default:
        return failure(404, 'NOT_FOUND')
    }
  }

  return {
    handle,
    seedUser: (user: TestUser = DEFAULT_USER): void => void addUser(user),
    seedSnippet: (seed: SnippetSeed): StoredSnippet => addSnippet(seed),
    requests: (): readonly string[] => [...requests],
    clearRequests: (): void => {
      requests = []
    },
    reset: (): void => {
      users = []
      snippets = []
      tokens = new Map()
      requests = []
    },
  }
}

export type FakeBackend = ReturnType<typeof createFakeBackend>
