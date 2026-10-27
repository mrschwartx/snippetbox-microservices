import { useState, type FormEvent } from 'react'
import { FormErrors } from '@/components/ui/FormErrors'
import { SubmitButton } from '@/components/ui/SubmitButton'
import { TextField } from '@/components/ui/TextField'
import { describeError } from '@/lib/describeError'
import { useSignup } from '../hooks/useSignup'

interface SignupFormProps {
  onSuccess: () => void
}

export function SignupForm({ onSuccess }: SignupFormProps) {
  const signup = useSignup()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const { fieldErrors, nonFieldErrors } = describeError(signup.error)

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    signup.mutate({ name, email, password }, { onSuccess })
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <FormErrors messages={nonFieldErrors} />
      <TextField
        id="name"
        label="Name:"
        value={name}
        onChange={setName}
        error={fieldErrors.name}
        autoComplete="name"
      />
      <TextField
        id="email"
        label="Email:"
        type="email"
        value={email}
        onChange={setEmail}
        error={fieldErrors.email}
        autoComplete="email"
      />
      <TextField
        id="password"
        label="Password:"
        type="password"
        value={password}
        onChange={setPassword}
        error={fieldErrors.password}
        autoComplete="new-password"
      />
      <SubmitButton label="Signup" isPending={signup.isPending} />
    </form>
  )
}
