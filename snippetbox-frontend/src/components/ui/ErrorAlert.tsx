interface ErrorAlertProps {
  message: string
  className?: string
}

export function ErrorAlert({ message, className = '' }: ErrorAlertProps) {
  return (
    <div role="alert" className={`bg-danger p-[1rem] text-center text-white ${className}`.trim()}>
      {message}
    </div>
  )
}
