import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { SiteHeader } from './SiteHeader'

describe('SiteHeader', () => {
  it('shows the brand as a link to the home page', () => {
    render(
      <MemoryRouter>
        <SiteHeader />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'SnippetBox' })).toHaveAttribute('href', '/')
  })
})
