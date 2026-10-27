import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FieldError } from './FieldError'

describe('FieldError', () => {
  it('renders nothing without a message', () => {
    const { container } = render(<FieldError id="title-error" />)

    expect(container).toBeEmptyDOMElement()
  })

  it('renders the message as an alert with the given id', () => {
    render(<FieldError id="title-error" message="This field cannot be blank" />)

    const alert = screen.getByRole('alert')
    expect(alert).toHaveTextContent('This field cannot be blank')
    expect(alert).toHaveAttribute('id', 'title-error')
  })
})
