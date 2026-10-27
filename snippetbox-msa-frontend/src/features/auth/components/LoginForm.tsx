import { useState, type FormEvent } from 'react'
import { FormErrors } from '@/components/ui/FormErrors'
import { SubmitButton } from '@/components/ui/SubmitButton'
import { TextField } from '@/components/ui/TextField'
import { describeError } from '@/lib/describeError'
import { useLogin } from '../hooks/useLogin'

interface LoginFormProps {
  onSuccess: () => void
}

export function LoginForm({ onSuccess }: LoginFormProps) {
  const login = useLogin()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const { fieldErrors, nonFieldErrors } = describeError(login.error)

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    login.mutate({ email, password }, { onSuccess })
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <FormErrors messages={nonFieldErrors} />
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
        autoComplete="current-password"
      />
      <SubmitButton label="Login" isPending={login.isPending} />
    </form>
  )
}
