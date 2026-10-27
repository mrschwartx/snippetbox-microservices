import { ErrorAlert } from './ErrorAlert'

interface FormErrorsProps {
  messages: string[]
}

export function FormErrors({ messages }: FormErrorsProps) {
  return (
    <>
      {messages.map((message) => (
        <ErrorAlert key={message} message={message} className="mb-9" />
      ))}
    </>
  )
}
