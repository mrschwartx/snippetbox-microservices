import { describe, expect, it } from 'vitest'
import { PASSWORD_MIN_LENGTH, validateLogin, validateSignup } from './validation'

const validSignup = { name: 'Alice', email: 'alice@example.com', password: 'correct-horse' }

describe('validateSignup', () => {
  it('accepts valid input', () => {
    expect(validateSignup(validSignup).valid).toBe(true)
  })

  it('reports every invalid field at once', () => {
    const { fieldErrors } = validateSignup({ name: '', email: 'nope', password: 'short' })

    expect(fieldErrors).toEqual({
      name: 'This field cannot be blank',
      email: 'This field must be a valid email address',
      password: 'This field must be at least 8 characters long',
    })
  })

  it('reports blank fields as blank rather than as malformed', () => {
    const { fieldErrors } = validateSignup({ name: ' ', email: '', password: '        ' })

    expect(fieldErrors).toEqual({
      name: 'This field cannot be blank',
      email: 'This field cannot be blank',
      password: 'This field cannot be blank',
    })
  })

  it('ignores surrounding whitespace around the email', () => {
    expect(validateSignup({ ...validSignup, email: '  alice@example.com  ' }).valid).toBe(true)
  })

  it('requires the minimum password length, counted in characters', () => {
    expect(
      validateSignup({ ...validSignup, password: 'a'.repeat(PASSWORD_MIN_LENGTH) }).valid,
    ).toBe(true)
    expect(
      validateSignup({ ...validSignup, password: 'a'.repeat(PASSWORD_MIN_LENGTH - 1) }).valid,
    ).toBe(false)
    expect(
      validateSignup({ ...validSignup, password: '😀'.repeat(PASSWORD_MIN_LENGTH) }).valid,
    ).toBe(true)
  })
})

describe('validateLogin', () => {
  it('accepts an email and a password', () => {
    expect(validateLogin({ email: 'a@b.co', password: 'x' }).valid).toBe(true)
  })

  it('requires both fields', () => {
    expect(validateLogin({ email: '', password: '' }).fieldErrors).toEqual({
      email: 'This field cannot be blank',
      password: 'This field cannot be blank',
    })
  })

  it('rejects a malformed email', () => {
    expect(validateLogin({ email: 'nope', password: 'x' }).fieldErrors).toEqual({
      email: 'This field must be a valid email address',
    })
  })

  it('does not impose a minimum password length on existing accounts', () => {
    expect(validateLogin({ email: 'a@b.co', password: 'x' }).fieldErrors.password).toBeUndefined()
  })
})
