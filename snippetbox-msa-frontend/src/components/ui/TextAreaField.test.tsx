import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { TextAreaField } from './TextAreaField'

function ControlledArea({ onChange }: { onChange: (value: string) => void }) {
  const [value, setValue] = useState('')

  return (
    <TextAreaField
      id="content"
      label="Content:"
      value={value}
      onChange={(next) => {
        setValue(next)
        onChange(next)
      }}
    />
  )
}

describe('TextAreaField', () => {
  it('labels its textarea', () => {
    render(<TextAreaField id="content" label="Content:" value="" onChange={vi.fn()} />)

    const textarea = screen.getByLabelText('Content:')
    expect(textarea.tagName).toBe('TEXTAREA')
    expect(textarea).toHaveAttribute('name', 'content')
  })

  it('reports multi-line input', async () => {
    const onChange = vi.fn()
    render(<ControlledArea onChange={onChange} />)

    await userEvent.setup().type(screen.getByLabelText('Content:'), 'a{Enter}b')

    expect(onChange).toHaveBeenLastCalledWith('a\nb')
  })

  it('marks the textarea invalid and links it to its error message', () => {
    render(
      <TextAreaField
        id="content"
        label="Content:"
        value=""
        onChange={vi.fn()}
        error="This field cannot be blank"
      />,
    )

    const textarea = screen.getByLabelText('Content:')
    expect(textarea).toHaveAttribute('aria-invalid', 'true')
    expect(textarea).toHaveAccessibleDescription('This field cannot be blank')
  })

  it('has no error state by default', () => {
    render(<TextAreaField id="content" label="Content:" value="" onChange={vi.fn()} />)

    expect(screen.getByLabelText('Content:')).not.toHaveAttribute('aria-invalid')
  })
})
