import { screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { createGate, signIn } from '@/test/helpers'
import { renderApp } from '@/test/render'
import { apiUrl, backend, server } from '@/test/server'

type User = ReturnType<typeof renderApp>['user']

const publish = async (user: User, title: string, content: string, expiry?: string) => {
  if (title) await user.type(await screen.findByLabelText('Title:'), title)
  if (content) await user.type(screen.getByLabelText('Content:'), content)
  if (expiry) await user.click(screen.getByLabelText(expiry))
  await user.click(screen.getByRole('button', { name: 'Publish snippet' }))
}

describe('SnippetCreatePage', () => {
  it('redirects an anonymous visitor to the login form', async () => {
    renderApp('/snippet/create')

    expect(await screen.findByRole('button', { name: 'Login' })).toBeInTheDocument()
    expect(screen.queryByLabelText('Title:')).not.toBeInTheDocument()
  })

  it('shows the form with a one year expiry preselected to a signed in user', async () => {
    await signIn()
    renderApp('/snippet/create')

    expect(await screen.findByLabelText('Title:')).toBeInTheDocument()
    expect(screen.getByLabelText('One Year')).toBeChecked()
    expect(screen.getByLabelText('One Week')).not.toBeChecked()
    expect(screen.getByLabelText('One Day')).not.toBeChecked()
  })

  it('offers the expiry options from longest to shortest', async () => {
    await signIn()
    renderApp('/snippet/create')

    await screen.findByLabelText('Title:')

    expect(screen.getAllByRole('radio').map((radio) => (radio as HTMLInputElement).value)).toEqual([
      '365',
      '7',
      '1',
    ])
  })

  it('publishes the snippet and shows it with a flash message', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')

    await publish(user, 'Hello world', 'fmt.Println("hi")', 'One Week')

    expect(await screen.findByText('Snippet successfully created!')).toBeInTheDocument()
    expect(screen.getByRole('article')).toHaveTextContent('Hello world')
    expect(screen.getByText('fmt.Println("hi")')).toBeInTheDocument()
    expect(screen.getByText('Expires: 22 Oct 2026 at 08:00')).toBeInTheDocument()
  })

  it.each([
    ['One Day', 'Expires: 16 Oct 2026 at 08:00'],
    ['One Week', 'Expires: 22 Oct 2026 at 08:00'],
    ['One Year', 'Expires: 15 Oct 2027 at 08:00'],
  ])('sends the expiry chosen with %s', async (expiry, expected) => {
    await signIn()
    const { user } = renderApp('/snippet/create')

    await publish(user, 'T', 'C', expiry)

    expect(await screen.findByText(expected)).toBeInTheDocument()
  })

  it('shows field errors, keeps the form and sends nothing when the input is invalid', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')
    await screen.findByLabelText('Title:')
    backend.clearRequests()

    await user.click(screen.getByRole('button', { name: 'Publish snippet' }))

    expect(await screen.findAllByText('This field cannot be blank')).toHaveLength(2)
    expect(screen.getByLabelText('Title:')).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByLabelText('One Year')).toBeChecked()
    expect(backend.requests()).toEqual([])
  })

  it('rejects a title longer than 100 characters', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')

    await publish(user, 'a'.repeat(101), 'body')

    expect(
      await screen.findByText('This field cannot be more than 100 characters long'),
    ).toBeInTheDocument()
  })

  it('disables the submit button while the request is in flight', async () => {
    await signIn()
    const gate = createGate()
    server.use(
      http.post(apiUrl('/snippets'), async () => {
        await gate.opened
        return HttpResponse.json({}, { status: 500 })
      }),
    )
    const { user } = renderApp('/snippet/create')

    await publish(user, 'T', 'C')

    expect(screen.getByRole('button', { name: 'Publish snippet' })).toBeDisabled()
    gate.open()
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Publish snippet' })).toBeEnabled(),
    )
  })

  it('keeps what was typed and shows a generic error when the server fails', async () => {
    await signIn()
    server.use(http.post(apiUrl('/snippets'), () => HttpResponse.json({}, { status: 500 })))
    const { user } = renderApp('/snippet/create')

    await publish(user, 'My title', 'My content')

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong')
    expect(screen.getByLabelText('Title:')).toHaveValue('My title')
    expect(screen.getByLabelText('Content:')).toHaveValue('My content')
  })

  it('removes the flash message on the next navigation', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')

    await publish(user, 'T', 'C')
    expect(await screen.findByText('Snippet successfully created!')).toBeInTheDocument()

    await user.click(screen.getByRole('link', { name: 'Home' }))

    await waitFor(() =>
      expect(screen.queryByText('Snippet successfully created!')).not.toBeInTheDocument(),
    )
    expect(await screen.findByRole('link', { name: 'T' })).toBeInTheDocument()
  })

  it('sends the user to the login form when the session is rejected by the server', async () => {
    await signIn()
    const { user } = renderApp('/snippet/create')
    await screen.findByLabelText('Title:')
    backend.reset()

    await publish(user, 'T', 'C')

    expect(await screen.findByRole('button', { name: 'Login' })).toBeInTheDocument()
    expect(screen.queryByLabelText('Title:')).not.toBeInTheDocument()
  })
})
