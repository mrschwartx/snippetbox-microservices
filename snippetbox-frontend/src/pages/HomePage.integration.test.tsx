import { screen } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { renderApp } from '@/test/render'
import { apiUrl, backend, server } from '@/test/server'
import { advanceHours } from '@/test/time'

describe('HomePage', () => {
  it('shows a loading indicator while the snippets are fetched', () => {
    renderApp('/')

    expect(screen.getByText('Loading…')).toBeInTheDocument()
  })

  it('shows an empty state when there are no snippets', async () => {
    renderApp('/')

    expect(await screen.findByText(/nothing to see here/i)).toBeInTheDocument()
  })

  it('lists the latest snippets, newest first, linking to each one', async () => {
    backend.seedSnippet({ title: 'First snippet' })
    backend.seedSnippet({ title: 'Second snippet' })
    renderApp('/')

    const links = await screen.findAllByRole('link', { name: /snippet$/ })

    expect(links.map((link) => link.textContent)).toEqual(['Second snippet', 'First snippet'])
    expect(links[0]).toHaveAttribute('href', '/snippet/view/2')
    expect(links[1]).toHaveAttribute('href', '/snippet/view/1')
  })

  it('leaves out snippets that have expired', async () => {
    backend.seedSnippet({ title: 'Gone tomorrow', days: 1 })
    backend.seedSnippet({ title: 'Here for a week', days: 7 })
    advanceHours(25)
    renderApp('/')

    expect(await screen.findByRole('link', { name: 'Here for a week' })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Gone tomorrow' })).not.toBeInTheDocument()
  })

  it('opens a snippet when its title is clicked', async () => {
    backend.seedSnippet({ title: 'Greeting', content: 'hello there' })
    const { user } = renderApp('/')

    await user.click(await screen.findByRole('link', { name: 'Greeting' }))

    expect(await screen.findByRole('article')).toHaveTextContent('hello there')
  })

  it('shows an error alert when the server fails', async () => {
    server.use(http.get(apiUrl('/snippets'), () => HttpResponse.json({}, { status: 500 })))
    renderApp('/')

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong')
  })

  it('shows an error alert when the network is down', async () => {
    server.use(http.get(apiUrl('/snippets'), () => HttpResponse.error()))
    renderApp('/')

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong')
  })
})
