import {
  InvalidCredentialsError,
  UNEXPECTED_ERROR_MESSAGE,
  UnexpectedError,
  ValidationError,
} from './errors'

export interface ErrorMessages {
  fieldErrors: Record<string, string>
  nonFieldErrors: string[]
}

export const describeError = (error: unknown): ErrorMessages => {
  if (error === null || error === undefined) return { fieldErrors: {}, nonFieldErrors: [] }

  if (error instanceof ValidationError) {
    return { fieldErrors: error.fieldErrors, nonFieldErrors: error.nonFieldErrors }
  }

  if (error instanceof InvalidCredentialsError || error instanceof UnexpectedError) {
    return { fieldErrors: {}, nonFieldErrors: [error.message] }
  }

  return { fieldErrors: {}, nonFieldErrors: [UNEXPECTED_ERROR_MESSAGE] }
}
