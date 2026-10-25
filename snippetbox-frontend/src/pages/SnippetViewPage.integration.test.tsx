import { screen, within } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { renderApp } from '@/test/render'
import { apiUrl, backend, server } from '@/test/server'
import { advanceHours } from '@/test/time'

describe('SnippetViewPage', () => {
  it('shows a loading indicator while the snippet is fetched', () => {
    backend.seedSnippet({ title: 'Greeting' })
    renderApp('/snippet/view/1')

    expect(screen.getByText('Loading…')).toBeInTheDocument()
  })

  it('shows title, content, id and dates', async () => {
    backend.seedSnippet({ title: 'Greeting', content: 'hello there', days: 1 })
    renderApp('/snippet/view/1')

    const article = await screen.findByRole('article')

    expect(within(article).getByText('Greeting')).toBeInTheDocument()
    expect(within(article).getByText('hello there')).toBeInTheDocument()
    expect(within(article).getByText('#1')).toBeInTheDocument()
    expect(within(article).getByText('Created: 15 Oct 2026 at 08:00')).toBeInTheDocument()
    expect(within(article).getByText('Expires: 16 Oct 2026 at 08:00')).toBeInTheDocument()
  })

  it('shows Not Found for an unknown id', async () => {
    renderApp('/snippet/view/99')

    expect(await screen.findByText('Not Found')).toBeInTheDocument()
    expect(screen.queryByRole('article')).not.toBeInTheDocument()
  })

  it.each(['abc', '1.5', '-1', '0'])(
    'shows Not Found for the malformed id %j without asking the server',
    async (id) => {
      renderApp(`/snippet/view/${id}`)

      expect(await screen.findByText('Not Found')).toBeInTheDocument()
      expect(backend.requests()).toEqual([])
    },
  )

  it('shows Not Found for a snippet that has expired', async () => {
    backend.seedSnippet({ title: 'Short lived', days: 1 })
    advanceHours(25)
    renderApp('/snippet/view/1')

    expect(await screen.findByText('Not Found')).toBeInTheDocument()
  })

  it('shows an error alert, not Not Found, when the server fails', async () => {
    backend.seedSnippet({ title: 'Greeting' })
    server.use(http.get(apiUrl('/snippets/1'), () => HttpResponse.json({}, { status: 500 })))
    renderApp('/snippet/view/1')

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong')
    expect(screen.queryByText('Not Found')).not.toBeInTheDocument()
  })
})
