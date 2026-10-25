import { inputClass } from './classNames'
import { getFieldAriaProps } from './fieldAria'
import { FormField } from './FormField'

interface TextFieldProps {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
  type?: 'text' | 'email' | 'password'
  error?: string
  autoComplete?: string
}

export function TextField({
  id,
  label,
  value,
  onChange,
  type = 'text',
  error,
  autoComplete,
}: TextFieldProps) {
  return (
    <FormField id={id} label={label} error={error}>
      <input
        id={id}
        name={id}
        type={type}
        value={value}
        autoComplete={autoComplete}
        onChange={(event) => onChange(event.target.value)}
        className={inputClass}
        {...getFieldAriaProps(id, error)}
      />
    </FormField>
  )
}
