import { screen } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { createGate } from '@/test/helpers'
import { renderApp } from '@/test/render'
import { apiUrl, backend, server } from '@/test/server'

type User = ReturnType<typeof renderApp>['user']

const fillAndSubmit = async (user: User, name: string, email: string, password: string) => {
  if (name) await user.type(screen.getByLabelText('Name:'), name)
  if (email) await user.type(screen.getByLabelText('Email:'), email)
  if (password) await user.type(screen.getByLabelText('Password:'), password)
  await user.click(screen.getByRole('button', { name: 'Signup' }))
}

describe('SignupPage', () => {
  it('shows an error for every blank field and does not call the server', async () => {
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, '', '', '')

    expect(await screen.findAllByText('This field cannot be blank')).toHaveLength(3)
    expect(backend.requests()).toEqual([])
  })

  it('rejects a malformed email', async () => {
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'not-an-email', 'correct-horse')

    expect(await screen.findByText('This field must be a valid email address')).toBeInTheDocument()
    expect(screen.getByLabelText('Email:')).toHaveAttribute('aria-invalid', 'true')
  })

  it('rejects a short password and marks the field invalid', async () => {
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'alice@example.com', 'short')

    expect(
      await screen.findByText('This field must be at least 8 characters long'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('Password:')).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByLabelText('Name:')).not.toHaveAttribute('aria-invalid')
  })

  it('shows the server message when the email is already taken', async () => {
    backend.seedUser()
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Eve', 'alice@example.com', 'another-pass')

    expect(await screen.findByText('Email address is already in use')).toBeInTheDocument()
    expect(screen.getByLabelText('Email:')).toHaveAttribute('aria-invalid', 'true')
  })

  it('redirects to the login form with a flash message after a successful signup', async () => {
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'alice@example.com', 'correct-horse')

    expect(
      await screen.findByText('Your signup was successful. Please log in.'),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument()
    expect(backend.requests()).toContain('POST /auth/signup')
  })

  it('disables the submit button while the request is in flight', async () => {
    const gate = createGate()
    server.use(
      http.post(apiUrl('/auth/signup'), async () => {
        await gate.opened
        return HttpResponse.json({ id: 1 }, { status: 201 })
      }),
    )
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'alice@example.com', 'correct-horse')

    expect(screen.getByRole('button', { name: 'Signup' })).toBeDisabled()
    gate.open()
    expect(await screen.findByRole('button', { name: 'Login' })).toBeInTheDocument()
  })

  it('shows a form-level error when the network is down and keeps what was typed', async () => {
    server.use(http.post(apiUrl('/auth/signup'), () => HttpResponse.error()))
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'alice@example.com', 'correct-horse')

    expect(
      await screen.findByText('Unable to reach the server. Please try again.'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('Email:')).toHaveValue('alice@example.com')
    expect(screen.getByRole('button', { name: 'Signup' })).toBeEnabled()
  })

  it('shows a generic error when the server fails', async () => {
    server.use(http.post(apiUrl('/auth/signup'), () => HttpResponse.json({}, { status: 500 })))
    const { user } = renderApp('/user/signup')

    await fillAndSubmit(user, 'Alice', 'alice@example.com', 'correct-horse')

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong')
  })
})
