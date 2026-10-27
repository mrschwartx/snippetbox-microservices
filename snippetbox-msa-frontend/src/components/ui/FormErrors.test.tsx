import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FormErrors } from './FormErrors'

describe('FormErrors', () => {
  it('renders nothing when there are no messages', () => {
    const { container } = render(<FormErrors messages={[]} />)

    expect(container).toBeEmptyDOMElement()
  })

  it('renders one alert per message', () => {
    render(<FormErrors messages={['First problem', 'Second problem']} />)

    expect(screen.getAllByRole('alert').map((alert) => alert.textContent)).toEqual([
      'First problem',
      'Second problem',
    ])
  })
})
