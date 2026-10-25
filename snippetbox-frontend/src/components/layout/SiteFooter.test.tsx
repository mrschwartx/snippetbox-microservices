import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { SiteFooter } from './SiteFooter'

describe('SiteFooter', () => {
  it('credits the technology and shows the current year', () => {
    render(<SiteFooter />)

    expect(screen.getByRole('contentinfo')).toHaveTextContent('Powered by Echo and React in 2026')
    expect(screen.getByRole('link', { name: 'Echo' })).toHaveAttribute(
      'href',
      'https://echo.labstack.com/',
    )
    expect(screen.getByRole('link', { name: 'React' })).toHaveAttribute(
      'href',
      'https://react.dev/',
    )
  })
})
