import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { SubmitButton } from './SubmitButton'

describe('SubmitButton', () => {
  it('is a submit button named after its label', () => {
    render(<SubmitButton label="Login" isPending={false} />)

    const button = screen.getByRole('button', { name: 'Login' })
    expect(button).toHaveAttribute('type', 'submit')
    expect(button).toBeEnabled()
  })

  it('is disabled while a submission is pending', () => {
    render(<SubmitButton label="Login" isPending />)

    expect(screen.getByRole('button', { name: 'Login' })).toBeDisabled()
  })
})
