import { EMAIL_PATTERN, Validator, matches, minChars, notBlank } from '@/lib/validator'
import type { LoginInput, SignupInput } from '../types'

export const PASSWORD_MIN_LENGTH = 8

export const validateSignup = (input: SignupInput): Validator => {
  const validator = new Validator()
  validator.check(notBlank(input.name), 'name', 'This field cannot be blank')
  validator.check(notBlank(input.email), 'email', 'This field cannot be blank')
  validator.check(
    matches(input.email.trim(), EMAIL_PATTERN),
    'email',
    'This field must be a valid email address',
  )
  validator.check(notBlank(input.password), 'password', 'This field cannot be blank')
  validator.check(
    minChars(input.password, PASSWORD_MIN_LENGTH),
    'password',
    `This field must be at least ${PASSWORD_MIN_LENGTH} characters long`,
  )
  return validator
}

export const validateLogin = (input: LoginInput): Validator => {
  const validator = new Validator()
  validator.check(notBlank(input.email), 'email', 'This field cannot be blank')
  validator.check(
    matches(input.email.trim(), EMAIL_PATTERN),
    'email',
    'This field must be a valid email address',
  )
  validator.check(notBlank(input.password), 'password', 'This field cannot be blank')
  return validator
}
