import { render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import type { Snippet } from '../types'
import { SnippetTable } from './SnippetTable'

const snippets: Snippet[] = [
  {
    id: 2,
    title: 'Second',
    content: 'b',
    created: '2026-10-15T09:15:00Z',
    expires: '2026-10-22T09:15:00Z',
  },
  {
    id: 1,
    title: 'First',
    content: 'a',
    created: '2026-10-14T07:05:00Z',
    expires: '2026-10-21T07:05:00Z',
  },
]

const renderTable = (rows: Snippet[]) =>
  render(
    <MemoryRouter>
      <SnippetTable snippets={rows} />
    </MemoryRouter>,
  )

describe('SnippetTable', () => {
  it('has Title, Created and ID columns', () => {
    renderTable(snippets)

    expect(screen.getAllByRole('columnheader').map((header) => header.textContent)).toEqual([
      'Title',
      'Created',
      'ID',
    ])
  })

  it('renders one row per snippet, in the order given', () => {
    renderTable(snippets)

    const [, ...rows] = screen.getAllByRole('row')
    expect(rows).toHaveLength(2)
    expect(within(rows[0]).getByText('Second')).toBeInTheDocument()
    expect(within(rows[1]).getByText('First')).toBeInTheDocument()
  })

  it('links each title to its snippet', () => {
    renderTable(snippets)

    expect(screen.getByRole('link', { name: 'Second' })).toHaveAttribute('href', '/snippet/view/2')
    expect(screen.getByRole('link', { name: 'First' })).toHaveAttribute('href', '/snippet/view/1')
  })

  it('shows the formatted creation date and the id of each snippet', () => {
    renderTable(snippets)

    const [, firstRow] = screen.getAllByRole('row')
    expect(within(firstRow).getByText('15 Oct 2026 at 09:15')).toBeInTheDocument()
    expect(within(firstRow).getByText('#2')).toBeInTheDocument()
  })

  it('renders only the header row for an empty list', () => {
    renderTable([])

    expect(screen.getAllByRole('row')).toHaveLength(1)
  })
})
