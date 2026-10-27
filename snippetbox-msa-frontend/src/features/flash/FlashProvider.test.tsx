import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Link, MemoryRouter, Route, Routes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { FlashMessage } from './FlashMessage'
import { FlashProvider } from './FlashProvider'
import { useFlash } from './useFlash'

function Queue() {
  const flash = useFlash()

  return <button onClick={() => flash.push('Saved!')}>Queue message</button>
}

const renderJourney = () =>
  render(
    <MemoryRouter initialEntries={['/a']}>
      <FlashProvider>
        <nav>
          <Link to="/a">A</Link>
          <Link to="/b">B</Link>
          <Link to="/c">C</Link>
        </nav>
        <FlashMessage />
        <Routes>
          <Route path="/a" element={<Queue />} />
          <Route path="/b" element={<p>Page B</p>} />
          <Route path="/c" element={<p>Page C</p>} />
        </Routes>
      </FlashProvider>
    </MemoryRouter>,
  )

describe('FlashProvider', () => {
  it('does not show a message on the page that queued it', async () => {
    renderJourney()

    await userEvent.setup().click(screen.getByRole('button', { name: 'Queue message' }))

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('shows the message on the next page the user lands on', async () => {
    const user = userEvent.setup()
    renderJourney()

    await user.click(screen.getByRole('button', { name: 'Queue message' }))
    await user.click(screen.getByRole('link', { name: 'B' }))

    expect(screen.getByRole('status')).toHaveTextContent('Saved!')
    expect(screen.getByText('Page B')).toBeInTheDocument()
  })

  it('drops the message as soon as the user navigates again', async () => {
    const user = userEvent.setup()
    renderJourney()

    await user.click(screen.getByRole('button', { name: 'Queue message' }))
    await user.click(screen.getByRole('link', { name: 'B' }))
    await user.click(screen.getByRole('link', { name: 'C' }))

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByText('Page C')).toBeInTheDocument()
  })

  it('does not bring the message back when the user returns to an earlier page', async () => {
    const user = userEvent.setup()
    renderJourney()

    await user.click(screen.getByRole('button', { name: 'Queue message' }))
    await user.click(screen.getByRole('link', { name: 'B' }))
    await user.click(screen.getByRole('link', { name: 'A' }))
    await user.click(screen.getByRole('link', { name: 'B' }))

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})
