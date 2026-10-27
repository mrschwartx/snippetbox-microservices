import { screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { sessionStore } from '@/lib/sessionStore'
import { DEFAULT_USER } from '@/test/fakeBackend'
import { createGate } from '@/test/helpers'
import { renderApp } from '@/test/render'
import { apiUrl, backend, server } from '@/test/server'

type User = ReturnType<typeof renderApp>['user']

const fillAndSubmit = async (user: User, email: string, password: string) => {
  if (email) await user.type(screen.getByLabelText('Email:'), email)
  if (password) await user.type(screen.getByLabelText('Password:'), password)
  await user.click(screen.getByRole('button', { name: 'Login' }))
}

describe('LoginPage', () => {
  it('requires an email and a password without calling the server', async () => {
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, '', '')

    expect(await screen.findAllByText('This field cannot be blank')).toHaveLength(2)
    expect(backend.requests()).toEqual([])
  })

  it('rejects a malformed email', async () => {
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, 'nope', 'secret')

    expect(await screen.findByText('This field must be a valid email address')).toBeInTheDocument()
  })

  it('shows a form-level error for wrong credentials and stores no session', async () => {
    backend.seedUser()
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, DEFAULT_USER.email, 'wrong-password')

    expect(await screen.findByText('Email or password is incorrect')).toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
  })

  it('lands on the create snippet form after a successful login', async () => {
    backend.seedUser()
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, DEFAULT_USER.email, DEFAULT_USER.password)

    expect(await screen.findByLabelText('Title:')).toBeInTheDocument()
    expect(sessionStore.read()?.token).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Logout' })).toBeInTheDocument()
  })

  it('accepts an email with surrounding whitespace', async () => {
    backend.seedUser()
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, `  ${DEFAULT_USER.email}  `, DEFAULT_USER.password)

    expect(await screen.findByLabelText('Title:')).toBeInTheDocument()
  })

  it('disables the submit button while the request is in flight', async () => {
    backend.seedUser()
    const gate = createGate()
    server.use(
      http.post(apiUrl('/auth/login'), async () => {
        await gate.opened
        return HttpResponse.json({ error: { code: 'INVALID_CREDENTIALS' } }, { status: 401 })
      }),
    )
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, DEFAULT_USER.email, DEFAULT_USER.password)

    expect(screen.getByRole('button', { name: 'Login' })).toBeDisabled()
    gate.open()
    await waitFor(() => expect(screen.getByRole('button', { name: 'Login' })).toBeEnabled())
    expect(screen.getByText('Email or password is incorrect')).toBeInTheDocument()
  })

  it('shows a form-level error when the network is down', async () => {
    server.use(http.post(apiUrl('/auth/login'), () => HttpResponse.error()))
    const { user } = renderApp('/user/login')

    await fillAndSubmit(user, DEFAULT_USER.email, DEFAULT_USER.password)

    expect(
      await screen.findByText('Unable to reach the server. Please try again.'),
    ).toBeInTheDocument()
  })
})
