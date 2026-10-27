import { describe, expect, it } from 'vitest'
import { ValidationError } from './errors'
import {
  EMAIL_PATTERN,
  Validator,
  matches,
  maxChars,
  minChars,
  notBlank,
  permittedValue,
} from './validator'

describe('Validator', () => {
  it('starts out valid', () => {
    const validator = new Validator()

    expect(validator.valid).toBe(true)
    expect(validator.fieldErrors).toEqual({})
  })

  it('ignores checks that pass', () => {
    const validator = new Validator()

    validator.check(true, 'title', 'never recorded')

    expect(validator.valid).toBe(true)
  })

  it('records a failed check against its field', () => {
    const validator = new Validator()

    validator.check(false, 'title', 'This field cannot be blank')

    expect(validator.valid).toBe(false)
    expect(validator.fieldErrors).toEqual({ title: 'This field cannot be blank' })
  })

  it('keeps only the first message for a field', () => {
    const validator = new Validator()

    validator.check(false, 'email', 'first')
    validator.check(false, 'email', 'second')

    expect(validator.fieldErrors).toEqual({ email: 'first' })
  })

  it('collects errors for several fields', () => {
    const validator = new Validator()

    validator.check(false, 'name', 'a')
    validator.check(false, 'email', 'b')

    expect(Object.keys(validator.fieldErrors)).toEqual(['name', 'email'])
  })

  it('assertValid does nothing when there are no errors', () => {
    expect(() => new Validator().assertValid()).not.toThrow()
  })

  it('assertValid throws a ValidationError carrying the recorded errors', () => {
    const validator = new Validator()
    validator.check(false, 'title', 'bad')

    expect(() => validator.assertValid()).toThrow(ValidationError)
    try {
      validator.assertValid()
    } catch (error) {
      expect((error as ValidationError).fieldErrors).toEqual({ title: 'bad' })
    }
  })
})

describe('notBlank', () => {
  it.each([
    ['', false],
    ['   ', false],
    ['\n\t', false],
    ['a', true],
    [' a ', true],
  ])('notBlank(%j) is %s', (value, expected) => {
    expect(notBlank(value)).toBe(expected)
  })
})

describe('maxChars and minChars', () => {
  it('count characters, not UTF-16 code units', () => {
    expect(maxChars('😀😀😀', 3)).toBe(true)
    expect(maxChars('😀😀😀😀', 3)).toBe(false)
    expect(minChars('😀😀😀', 3)).toBe(true)
    expect(minChars('😀😀', 3)).toBe(false)
  })

  it('treat the limit as inclusive', () => {
    expect(maxChars('abc', 3)).toBe(true)
    expect(maxChars('abcd', 3)).toBe(false)
    expect(minChars('abc', 3)).toBe(true)
    expect(minChars('ab', 3)).toBe(false)
  })
})

describe('permittedValue', () => {
  it('accepts only listed values', () => {
    expect(permittedValue(7, [1, 7, 365])).toBe(true)
    expect(permittedValue(2, [1, 7, 365])).toBe(false)
    expect(permittedValue(Number.NaN, [1, 7, 365])).toBe(false)
  })
})

describe('matches', () => {
  it('tests a value against a pattern', () => {
    expect(matches('abc', /^a/)).toBe(true)
    expect(matches('xbc', /^a/)).toBe(false)
  })
})

describe('EMAIL_PATTERN', () => {
  it.each([
    'a@b.co',
    'first.last+tag@sub.example.com',
    "o'brien@example.com",
    'user@localhost',
    'UPPER@EXAMPLE.COM',
  ])('accepts %s', (email) => {
    expect(EMAIL_PATTERN.test(email)).toBe(true)
  })

  it.each([
    '',
    'missing-at.example.com',
    'a@',
    '@b.co',
    'a b@c.co',
    'a@-b.co',
    'a@b..co',
    'a@b.co ',
    'a@@b.co',
  ])('rejects %j', (email) => {
    expect(EMAIL_PATTERN.test(email)).toBe(false)
  })
})
