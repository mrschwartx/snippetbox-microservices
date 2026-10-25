import { apiClient } from '@/lib/apiClient'
import { UnauthenticatedError } from '@/lib/errors'
import { isSessionActive, sessionStore } from '@/lib/sessionStore'
import { validateLogin, validateSignup } from '../lib/validation'
import type { LoginInput, SignupInput, User } from '../types'

interface LoginResponse {
  token: string
  token_type: string
  expires_at: string
  user: User
}

export const signup = async (input: SignupInput): Promise<User> => {
  validateSignup(input).assertValid()

  return apiClient.post<User>('/auth/signup', {
    body: { name: input.name.trim(), email: input.email.trim(), password: input.password },
  })
}

export const login = async (input: LoginInput): Promise<User> => {
  validateLogin(input).assertValid()

  const response = await apiClient.post<LoginResponse>('/auth/login', {
    body: { email: input.email.trim(), password: input.password },
  })
  sessionStore.write({ token: response.token, expiresAt: response.expires_at })
  return response.user
}

export const logout = async (): Promise<void> => {
  await apiClient.post<void>('/auth/logout', { auth: true }).catch(() => undefined)
  sessionStore.clear()
}

export const getCurrentUser = async (): Promise<User | null> => {
  const session = sessionStore.read()
  if (!session) return null

  if (!isSessionActive(session)) {
    sessionStore.clear()
    return null
  }

  try {
    return await apiClient.get<User>('/auth/me', { auth: true })
  } catch (error) {
    if (error instanceof UnauthenticatedError) {
      sessionStore.clear()
      return null
    }
    throw error
  }
}
