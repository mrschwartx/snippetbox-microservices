import { inputClass } from './classNames'
import { getFieldAriaProps } from './fieldAria'
import { FormField } from './FormField'

interface TextAreaFieldProps {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
  error?: string
}

export function TextAreaField({ id, label, value, onChange, error }: TextAreaFieldProps) {
  return (
    <FormField id={id} label={label} error={error}>
      <textarea
        id={id}
        name={id}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`${inputClass} h-[266px] resize-y p-[1rem]`}
        {...getFieldAriaProps(id, error)}
      />
    </FormField>
  )
}
