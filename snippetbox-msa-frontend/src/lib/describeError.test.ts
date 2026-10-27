import { describe, expect, it } from 'vitest'
import { describeError } from './describeError'
import {
  InvalidCredentialsError,
  NoRecordError,
  UNEXPECTED_ERROR_MESSAGE,
  UnauthenticatedError,
  UnexpectedError,
  ValidationError,
} from './errors'

describe('describeError', () => {
  it.each([null, undefined])('returns no messages for %s', (error) => {
    expect(describeError(error)).toEqual({ fieldErrors: {}, nonFieldErrors: [] })
  })

  it('exposes the field and non-field errors of a ValidationError', () => {
    const error = new ValidationError({ email: 'taken' }, ['try again'])

    expect(describeError(error)).toEqual({
      fieldErrors: { email: 'taken' },
      nonFieldErrors: ['try again'],
    })
  })

  it('shows the message of invalid credentials as a form-level error', () => {
    expect(describeError(new InvalidCredentialsError())).toEqual({
      fieldErrors: {},
      nonFieldErrors: ['Email or password is incorrect'],
    })
  })

  it('shows the message of an UnexpectedError, such as a network failure', () => {
    expect(describeError(new UnexpectedError('Unable to reach the server.'))).toEqual({
      fieldErrors: {},
      nonFieldErrors: ['Unable to reach the server.'],
    })
  })

  it.each([
    ['a plain Error', new Error('boom')],
    ['a NoRecordError', new NoRecordError()],
    ['an UnauthenticatedError', new UnauthenticatedError()],
    ['a string', 'boom'],
  ])('falls back to a generic message for %s', (_label, error) => {
    expect(describeError(error)).toEqual({
      fieldErrors: {},
      nonFieldErrors: [UNEXPECTED_ERROR_MESSAGE],
    })
  })
})
