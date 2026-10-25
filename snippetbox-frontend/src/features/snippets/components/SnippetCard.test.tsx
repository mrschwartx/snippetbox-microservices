import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { SnippetCard } from './SnippetCard'

const snippet = {
  id: 7,
  title: 'Greeting',
  content: 'hello there',
  created: '2026-10-15T08:00:00Z',
  expires: '2026-10-22T08:30:00Z',
}

describe('SnippetCard', () => {
  it('shows the title and id', () => {
    render(<SnippetCard snippet={snippet} />)

    const article = screen.getByRole('article')
    expect(within(article).getByText('Greeting')).toBeInTheDocument()
    expect(within(article).getByText('#7')).toBeInTheDocument()
  })

  it('shows the creation and expiry dates in a readable form', () => {
    render(<SnippetCard snippet={snippet} />)

    expect(screen.getByText('Created: 15 Oct 2026 at 08:00')).toHaveAttribute(
      'datetime',
      '2026-10-15T08:00:00Z',
    )
    expect(screen.getByText('Expires: 22 Oct 2026 at 08:30')).toHaveAttribute(
      'datetime',
      '2026-10-22T08:30:00Z',
    )
  })

  it('preserves the whitespace of the content', () => {
    render(<SnippetCard snippet={{ ...snippet, content: 'if x {\n    run()\n}' }} />)

    expect(screen.getByText(/run\(\)/).textContent).toBe('if x {\n    run()\n}')
    expect(screen.getByText(/run\(\)/).tagName).toBe('CODE')
  })

  it('renders markup in the content as plain text', () => {
    const { container } = render(
      <SnippetCard snippet={{ ...snippet, content: '<script>alert(1)</script>' }} />,
    )

    expect(screen.getByText('<script>alert(1)</script>')).toBeInTheDocument()
    expect(container.querySelector('script')).toBeNull()
  })
})
