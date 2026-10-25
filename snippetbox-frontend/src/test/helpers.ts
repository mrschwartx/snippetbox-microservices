import { login } from '@/features/auth/api/authApi'
import { DEFAULT_USER, type TestUser } from './fakeBackend'
import { backend } from './server'

export const signIn = async (user: TestUser = DEFAULT_USER): Promise<void> => {
  backend.seedUser(user)
  await login({ email: user.email, password: user.password })
}

export const createGate = () => {
  let open!: () => void
  const opened = new Promise<void>((resolve) => {
    open = resolve
  })
  return { opened, open }
}
