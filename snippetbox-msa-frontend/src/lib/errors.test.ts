import { describe, expect, it } from 'vitest'
import {
  AppError,
  InvalidCredentialsError,
  NoRecordError,
  UNEXPECTED_ERROR_MESSAGE,
  UnauthenticatedError,
  UnexpectedError,
  ValidationError,
} from './errors'

describe('errors', () => {
  it.each([
    [new NoRecordError(), 'NoRecordError', 'NO_RECORD', 'No matching record found'],
    [
      new InvalidCredentialsError(),
      'InvalidCredentialsError',
      'INVALID_CREDENTIALS',
      'Email or password is incorrect',
    ],
    [
      new UnauthenticatedError(),
      'UnauthenticatedError',
      'UNAUTHENTICATED',
      'Authentication required',
    ],
    [new ValidationError(), 'ValidationError', 'VALIDATION', 'Validation failed'],
    [new UnexpectedError(), 'UnexpectedError', 'UNEXPECTED', UNEXPECTED_ERROR_MESSAGE],
  ])('%o has a stable name, code and message', (error, name, code, message) => {
    expect(error).toBeInstanceOf(AppError)
    expect(error).toBeInstanceOf(Error)
    expect(error.name).toBe(name)
    expect(error.code).toBe(code)
    expect(error.message).toBe(message)
  })

  it('ValidationError defaults to empty field and non-field errors', () => {
    const error = new ValidationError()

    expect(error.fieldErrors).toEqual({})
    expect(error.nonFieldErrors).toEqual([])
  })

  it('ValidationError carries the errors it was given', () => {
    const error = new ValidationError({ title: 'bad' }, ['oops'])

    expect(error.fieldErrors).toEqual({ title: 'bad' })
    expect(error.nonFieldErrors).toEqual(['oops'])
  })

  it('UnexpectedError accepts a custom message', () => {
    expect(new UnexpectedError('server is down').message).toBe('server is down')
  })

  it('errors of different kinds are distinguishable with instanceof', () => {
    expect(new NoRecordError()).not.toBeInstanceOf(UnauthenticatedError)
    expect(new UnauthenticatedError()).not.toBeInstanceOf(InvalidCredentialsError)
  })
})
