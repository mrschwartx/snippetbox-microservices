import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ErrorAlert } from './ErrorAlert'

describe('ErrorAlert', () => {
  it('announces its message as an alert', () => {
    render(<ErrorAlert message="Something went wrong" />)

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong')
  })

  it('accepts extra classes without losing its own styling', () => {
    render(<ErrorAlert message="Oops" className="mb-9" />)

    expect(screen.getByRole('alert')).toHaveClass('bg-danger', 'text-white', 'mb-9')
  })
})
