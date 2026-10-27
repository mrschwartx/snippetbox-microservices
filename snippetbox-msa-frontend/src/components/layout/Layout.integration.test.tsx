import { screen } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { sessionStore } from '@/lib/sessionStore'
import { createGate, signIn } from '@/test/helpers'
import { renderApp } from '@/test/render'
import { apiUrl, server } from '@/test/server'

describe('Layout', () => {
  it('shows the brand, the navigation and the footer on every page', async () => {
    renderApp('/nowhere')

    expect(screen.getByRole('link', { name: 'SnippetBox' })).toHaveAttribute('href', '/')
    expect(screen.getByRole('navigation')).toBeInTheDocument()
    expect(screen.getByRole('contentinfo')).toBeInTheDocument()
    expect(await screen.findByText('Not Found')).toBeInTheDocument()
  })

  it('offers a visitor Home, Signup and Login only', async () => {
    renderApp('/')

    expect(await screen.findByRole('link', { name: 'Signup' })).toHaveAttribute(
      'href',
      '/user/signup',
    )
    expect(screen.getByRole('link', { name: 'Login' })).toHaveAttribute('href', '/user/login')
    expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Create snippet' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Logout' })).not.toBeInTheDocument()
  })

  it('offers a signed in user Create snippet and Logout instead', async () => {
    await signIn()
    renderApp('/')

    expect(await screen.findByRole('button', { name: 'Logout' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Create snippet' })).toHaveAttribute(
      'href',
      '/snippet/create',
    )
    expect(screen.queryByRole('link', { name: 'Login' })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Signup' })).not.toBeInTheDocument()
  })

  it('shows neither visitor nor user links while the session is being resolved', async () => {
    await signIn()
    const gate = createGate()
    server.use(
      http.get(apiUrl('/auth/me'), async () => {
        await gate.opened
        return HttpResponse.json({ id: 1, name: 'Alice', email: 'a@b.co', created: 'x' })
      }),
    )
    renderApp('/')

    expect(screen.queryByRole('link', { name: 'Login' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Logout' })).not.toBeInTheDocument()

    gate.open()
    expect(await screen.findByRole('button', { name: 'Logout' })).toBeInTheDocument()
  })

  it('marks the link of the current page as current', async () => {
    renderApp('/user/login')

    expect(await screen.findByRole('link', { name: 'Login' })).toHaveAttribute(
      'aria-current',
      'page',
    )
    expect(screen.getByRole('link', { name: 'Signup' })).not.toHaveAttribute('aria-current')
  })

  it('logs the user out, clears the session and says so', async () => {
    await signIn()
    const { user } = renderApp('/')

    await user.click(await screen.findByRole('button', { name: 'Logout' }))

    expect(await screen.findByText("You've been logged out successfully!")).toBeInTheDocument()
    expect(await screen.findByRole('link', { name: 'Login' })).toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
  })

  it('returns to the home page when the user logs out from a protected page', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')

    await user.click(await screen.findByRole('button', { name: 'Logout' }))

    expect(await screen.findByText('Latest Snippets')).toBeInTheDocument()
    expect(screen.getByText("You've been logged out successfully!")).toBeInTheDocument()
    expect(screen.queryByLabelText('Password:')).not.toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
  })

  it('still logs the user out locally when the server cannot be reached', async () => {
    await signIn()
    server.use(http.post(apiUrl('/auth/logout'), () => HttpResponse.error()))
    const { user } = renderApp('/')

    await user.click(await screen.findByRole('button', { name: 'Logout' }))

    expect(await screen.findByRole('link', { name: 'Login' })).toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
  })
})
