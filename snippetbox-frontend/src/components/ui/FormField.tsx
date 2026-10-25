import type { ReactNode } from 'react'
import { FieldError } from './FieldError'

interface FormFieldProps {
  id: string
  label: string
  error?: string
  children: ReactNode
}

export function FormField({ id, label, error, children }: FormFieldProps) {
  return (
    <div className="mb-[1rem]">
      <label htmlFor={id} className="mb-2 inline-block">
        {label}
      </label>
      <FieldError id={`${id}-error`} message={error} />
      {children}
    </div>
  )
}
