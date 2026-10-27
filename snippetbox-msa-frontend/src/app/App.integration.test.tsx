import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { sessionStore } from '@/lib/sessionStore'
import { DEFAULT_USER } from '@/test/fakeBackend'
import { signIn } from '@/test/helpers'
import { renderApp } from '@/test/render'
import { backend } from '@/test/server'
import { advanceHours } from '@/test/time'

describe('App', () => {
  it('lets a visitor sign up, log in, publish a snippet and find it on the home page', async () => {
    const { user } = renderApp('/')

    await user.click(await screen.findByRole('link', { name: 'Signup' }))
    await user.type(screen.getByLabelText('Name:'), DEFAULT_USER.name)
    await user.type(screen.getByLabelText('Email:'), DEFAULT_USER.email)
    await user.type(screen.getByLabelText('Password:'), DEFAULT_USER.password)
    await user.click(screen.getByRole('button', { name: 'Signup' }))
    expect(
      await screen.findByText('Your signup was successful. Please log in.'),
    ).toBeInTheDocument()

    await user.type(screen.getByLabelText('Email:'), DEFAULT_USER.email)
    await user.type(screen.getByLabelText('Password:'), DEFAULT_USER.password)
    await user.click(screen.getByRole('button', { name: 'Login' }))

    await user.type(await screen.findByLabelText('Title:'), 'My first snippet')
    await user.type(screen.getByLabelText('Content:'), 'console.log("hi")')
    await user.click(screen.getByRole('button', { name: 'Publish snippet' }))
    expect(await screen.findByText('Snippet successfully created!')).toBeInTheDocument()
    expect(screen.getByRole('article')).toHaveTextContent('My first snippet')

    await user.click(screen.getByRole('link', { name: 'Home' }))
    expect(await screen.findByRole('link', { name: 'My first snippet' })).toHaveAttribute(
      'href',
      '/snippet/view/1',
    )
  })

  it('keeps the user signed in across a page reload', async () => {
    await signIn()

    const first = renderApp('/')
    expect(await screen.findByRole('button', { name: 'Logout' })).toBeInTheDocument()
    first.unmount()

    renderApp('/')
    expect(await screen.findByRole('button', { name: 'Logout' })).toBeInTheDocument()
  })

  it('treats a session that expired while the app was closed as signed out', async () => {
    await signIn()
    advanceHours(13)
    backend.clearRequests()

    renderApp('/snippet/create')

    expect(await screen.findByRole('button', { name: 'Login' })).toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
    expect(backend.requests()).toEqual([])
  })

  it('treats a session the server revoked as signed out', async () => {
    await signIn()
    backend.reset()

    renderApp('/')

    expect(await screen.findByRole('link', { name: 'Login' })).toBeInTheDocument()
    expect(sessionStore.read()).toBeNull()
  })

  it('shows Not Found inside the layout for an unknown route', async () => {
    renderApp('/definitely/not/here')

    expect(await screen.findByText('Not Found')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument()
  })

  it('lets a signed in user reach the create form from the navigation', async () => {
    await signIn()
    const { user } = renderApp('/')

    await user.click(await screen.findByRole('link', { name: 'Create snippet' }))

    expect(await screen.findByLabelText('Title:')).toBeInTheDocument()
  })
})
