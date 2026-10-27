export const UNEXPECTED_ERROR_MESSAGE = 'Something went wrong. Please try again.'

export class AppError extends Error {
  readonly code: string

  constructor(code: string, message: string) {
    super(message)
    this.code = code
    this.name = new.target.name
  }
}

export class NoRecordError extends AppError {
  constructor() {
    super('NO_RECORD', 'No matching record found')
  }
}

export class InvalidCredentialsError extends AppError {
  constructor() {
    super('INVALID_CREDENTIALS', 'Email or password is incorrect')
  }
}

export class UnauthenticatedError extends AppError {
  constructor() {
    super('UNAUTHENTICATED', 'Authentication required')
  }
}

export class ValidationError extends AppError {
  readonly fieldErrors: Record<string, string>
  readonly nonFieldErrors: string[]

  constructor(fieldErrors: Record<string, string> = {}, nonFieldErrors: string[] = []) {
    super('VALIDATION', 'Validation failed')
    this.fieldErrors = fieldErrors
    this.nonFieldErrors = nonFieldErrors
  }
}

export class UnexpectedError extends AppError {
  constructor(message = UNEXPECTED_ERROR_MESSAGE) {
    super('UNEXPECTED', message)
  }
}
