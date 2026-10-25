import { describe, expect, it } from 'vitest'
import { EXPIRY_OPTIONS, PERMITTED_EXPIRES, TITLE_MAX_LENGTH } from '../constants'
import { validateSnippetDraft } from './validation'

const validDraft = { title: 'Hello', content: 'body', expires: 7 }

describe('validateSnippetDraft', () => {
  it('accepts a valid draft', () => {
    expect(validateSnippetDraft(validDraft).valid).toBe(true)
  })

  it('reports blank fields and an invalid expiry together', () => {
    expect(validateSnippetDraft({ title: ' ', content: '', expires: 3 }).fieldErrors).toEqual({
      title: 'This field cannot be blank',
      content: 'This field cannot be blank',
      expires: 'This field must equal 1, 7 or 365',
    })
  })

  it('limits the title to 100 characters, counted as characters', () => {
    const tooLong = validateSnippetDraft({
      ...validDraft,
      title: 'a'.repeat(TITLE_MAX_LENGTH + 1),
    })

    expect(tooLong.fieldErrors.title).toBe('This field cannot be more than 100 characters long')
    expect(
      validateSnippetDraft({ ...validDraft, title: '😀'.repeat(TITLE_MAX_LENGTH) }).valid,
    ).toBe(true)
    expect(validateSnippetDraft({ ...validDraft, title: 'a'.repeat(TITLE_MAX_LENGTH) }).valid).toBe(
      true,
    )
  })

  it('reports a blank title as blank rather than as too long', () => {
    expect(validateSnippetDraft({ ...validDraft, title: '' }).fieldErrors.title).toBe(
      'This field cannot be blank',
    )
  })

  it.each([1, 7, 365])('permits an expiry of %i days', (expires) => {
    expect(validateSnippetDraft({ ...validDraft, expires }).valid).toBe(true)
  })

  it.each([0, 2, 30, 366, -1, 7.5, Number.NaN])('rejects an expiry of %s days', (expires) => {
    expect(validateSnippetDraft({ ...validDraft, expires }).fieldErrors.expires).toBe(
      'This field must equal 1, 7 or 365',
    )
  })
})

describe('snippet constants', () => {
  it('permits exactly the expiries named in the validation message', () => {
    expect([...PERMITTED_EXPIRES].sort((a, b) => a - b)).toEqual([1, 7, 365])
  })

  it('offers the longest expiry first', () => {
    expect(EXPIRY_OPTIONS.map((option) => option.label)).toEqual([
      'One Year',
      'One Week',
      'One Day',
    ])
  })
})
