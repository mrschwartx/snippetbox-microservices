export interface Session {
  token: string
  expiresAt: string
}

export interface SessionStore {
  read(): Session | null
  write(session: Session): void
  clear(): void
}

type StorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>

export const SESSION_STORAGE_KEY = 'snippetbox:session'
export const SESSION_EXPIRY_SKEW_MS = 5_000

const isSession = (value: unknown): value is Session =>
  typeof value === 'object' &&
  value !== null &&
  typeof (value as Session).token === 'string' &&
  typeof (value as Session).expiresAt === 'string'

export const createSessionStore = (
  resolveStorage: () => StorageLike = () => window.localStorage,
): SessionStore => ({
  read() {
    try {
      const raw = resolveStorage().getItem(SESSION_STORAGE_KEY)
      if (!raw) return null
      const parsed: unknown = JSON.parse(raw)
      return isSession(parsed) ? parsed : null
    } catch {
      return null
    }
  },

  write(session) {
    try {
      resolveStorage().setItem(SESSION_STORAGE_KEY, JSON.stringify(session))
    } catch {
      return
    }
  },

  clear() {
    try {
      resolveStorage().removeItem(SESSION_STORAGE_KEY)
    } catch {
      return
    }
  },
})

export const sessionStore = createSessionStore()

export const isSessionActive = (session: Session, now: Date = new Date()): boolean =>
  new Date(session.expiresAt).getTime() - SESSION_EXPIRY_SKEW_MS > now.getTime()
