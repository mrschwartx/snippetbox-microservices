import { useState, type FormEvent } from 'react'
import { FieldError } from '@/components/ui/FieldError'
import { FormErrors } from '@/components/ui/FormErrors'
import { SubmitButton } from '@/components/ui/SubmitButton'
import { TextAreaField } from '@/components/ui/TextAreaField'
import { TextField } from '@/components/ui/TextField'
import { describeError } from '@/lib/describeError'
import { DEFAULT_EXPIRES_DAYS, EXPIRY_OPTIONS } from '../constants'
import { useCreateSnippet } from '../hooks/useCreateSnippet'
import type { Snippet } from '../types'

interface SnippetFormProps {
  onCreated: (snippet: Snippet) => void
}

export function SnippetForm({ onCreated }: SnippetFormProps) {
  const createSnippet = useCreateSnippet()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [expires, setExpires] = useState<number>(DEFAULT_EXPIRES_DAYS)
  const { fieldErrors, nonFieldErrors } = describeError(createSnippet.error)

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    createSnippet.mutate({ title, content, expires }, { onSuccess: onCreated })
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <FormErrors messages={nonFieldErrors} />
      <TextField
        id="title"
        label="Title:"
        value={title}
        onChange={setTitle}
        error={fieldErrors.title}
      />
      <TextAreaField
        id="content"
        label="Content:"
        value={content}
        onChange={setContent}
        error={fieldErrors.content}
      />
      <fieldset className="mb-[1rem]">
        <legend className="mb-2">Delete in:</legend>
        <FieldError id="expires-error" message={fieldErrors.expires} />
        {EXPIRY_OPTIONS.map(({ days, label }) => (
          <label key={days} className="mr-[1rem] inline-block">
            <input
              type="radio"
              name="expires"
              value={days}
              checked={expires === days}
              onChange={() => setExpires(days)}
            />{' '}
            {label}
          </label>
        ))}
      </fieldset>
      <SubmitButton label="Publish snippet" isPending={createSnippet.isPending} />
    </form>
  )
}
