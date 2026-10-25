import { Validator, maxChars, notBlank, permittedValue } from '@/lib/validator'
import { PERMITTED_EXPIRES, TITLE_MAX_LENGTH } from '../constants'
import type { SnippetDraft } from '../types'

export const validateSnippetDraft = (draft: SnippetDraft): Validator => {
  const validator = new Validator()
  validator.check(notBlank(draft.title), 'title', 'This field cannot be blank')
  validator.check(
    maxChars(draft.title, TITLE_MAX_LENGTH),
    'title',
    `This field cannot be more than ${TITLE_MAX_LENGTH} characters long`,
  )
  validator.check(notBlank(draft.content), 'content', 'This field cannot be blank')
  validator.check(
    permittedValue(draft.expires, PERMITTED_EXPIRES),
    'expires',
    'This field must equal 1, 7 or 365',
  )
  return validator
}
