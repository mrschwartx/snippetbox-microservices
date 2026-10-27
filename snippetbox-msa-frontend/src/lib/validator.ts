import { ValidationError } from './errors'

export const EMAIL_PATTERN =
  /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/

export class Validator {
  readonly fieldErrors: Record<string, string> = {}
  readonly nonFieldErrors: string[] = []

  get valid(): boolean {
    return Object.keys(this.fieldErrors).length === 0 && this.nonFieldErrors.length === 0
  }

  check(ok: boolean, field: string, message: string): void {
    if (!ok && !(field in this.fieldErrors)) {
      this.fieldErrors[field] = message
    }
  }

  assertValid(): void {
    if (!this.valid) throw new ValidationError(this.fieldErrors, this.nonFieldErrors)
  }
}

export const notBlank = (value: string): boolean => value.trim() !== ''

export const maxChars = (value: string, limit: number): boolean => Array.from(value).length <= limit

export const minChars = (value: string, limit: number): boolean => Array.from(value).length >= limit

export const matches = (value: string, pattern: RegExp): boolean => pattern.test(value)

export const permittedValue = <T>(value: T, permitted: readonly T[]): boolean =>
  permitted.includes(value)
