import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Loading } from './Loading'

describe('Loading', () => {
  it('tells the user something is loading', () => {
    render(<Loading />)

    expect(screen.getByText('Loading…')).toBeInTheDocument()
  })
})
