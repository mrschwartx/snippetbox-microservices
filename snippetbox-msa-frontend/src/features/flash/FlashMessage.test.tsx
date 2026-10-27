import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { FlashContext } from './FlashContext'
import { FlashMessage } from './FlashMessage'

const renderWithMessage = (message: string | null) =>
  render(
    <FlashContext.Provider value={{ message, push: vi.fn() }}>
      <FlashMessage />
    </FlashContext.Provider>,
  )

describe('FlashMessage', () => {
  it('renders nothing when there is no message', () => {
    const { container } = renderWithMessage(null)

    expect(container).toBeEmptyDOMElement()
  })

  it('announces the message as a status update', () => {
    renderWithMessage('Saved!')

    expect(screen.getByRole('status')).toHaveTextContent('Saved!')
  })
})
