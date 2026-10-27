interface FieldErrorProps {
  id: string
  message?: string
}

export function FieldError({ id, message }: FieldErrorProps) {
  if (!message) return null

  return (
    <p id={id} role="alert" className="block font-bold text-danger">
      {message}
    </p>
  )
}
