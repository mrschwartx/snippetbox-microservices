import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { TextField } from './TextField'

function ControlledField({ onChange }: { onChange?: (value: string) => void }) {
  const [value, setValue] = useState('')

  return (
    <TextField
      id="name"
      label="Name:"
      value={value}
      onChange={(next) => {
        setValue(next)
        onChange?.(next)
      }}
    />
  )
}

describe('TextField', () => {
  it('labels its input', () => {
    render(<TextField id="name" label="Name:" value="" onChange={vi.fn()} />)

    expect(screen.getByLabelText('Name:')).toHaveAttribute('name', 'name')
  })

  it('is a text input unless told otherwise', () => {
    render(<TextField id="name" label="Name:" value="" onChange={vi.fn()} />)

    expect(screen.getByLabelText('Name:')).toHaveAttribute('type', 'text')
  })

  it.each(['email', 'password'] as const)('supports a %s input', (type) => {
    render(<TextField id="field" label="Field:" type={type} value="" onChange={vi.fn()} />)

    expect(screen.getByLabelText('Field:')).toHaveAttribute('type', type)
  })

  it('passes the autocomplete hint to the input', () => {
    render(<TextField id="email" label="Email:" value="" onChange={vi.fn()} autoComplete="email" />)

    expect(screen.getByLabelText('Email:')).toHaveAttribute('autocomplete', 'email')
  })

  it('shows the value it is given', () => {
    render(<TextField id="name" label="Name:" value="Alice" onChange={vi.fn()} />)

    expect(screen.getByLabelText('Name:')).toHaveValue('Alice')
  })

  it('reports the new value as the user types', async () => {
    const onChange = vi.fn()
    render(<ControlledField onChange={onChange} />)

    await userEvent.setup().type(screen.getByLabelText('Name:'), 'Bob')

    expect(onChange.mock.calls.map(([value]) => value)).toEqual(['B', 'Bo', 'Bob'])
  })

  it('has no error state by default', () => {
    render(<TextField id="name" label="Name:" value="" onChange={vi.fn()} />)

    const input = screen.getByLabelText('Name:')
    expect(input).not.toHaveAttribute('aria-invalid')
    expect(input).not.toHaveAttribute('aria-describedby')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('marks the input invalid and links it to its error message', () => {
    render(
      <TextField
        id="name"
        label="Name:"
        value=""
        onChange={vi.fn()}
        error="This field cannot be blank"
      />,
    )

    const input = screen.getByLabelText('Name:')
    expect(input).toHaveAttribute('aria-invalid', 'true')
    expect(input).toHaveAccessibleDescription('This field cannot be blank')
    expect(screen.getByRole('alert')).toHaveTextContent('This field cannot be blank')
  })
})
