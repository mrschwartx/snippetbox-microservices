import { describe, expect, it } from 'vitest'
import { parseSnippetId } from './parseSnippetId'

describe('parseSnippetId', () => {
  it.each([
    ['1', 1],
    ['42', 42],
    ['007', 7],
    ['0', 0],
  ])('parses %j as %i', (param, expected) => {
    expect(parseSnippetId(param)).toBe(expected)
  })

  it.each(['abc', '1.5', '-1', '', ' 12', '12abc', '1e3', undefined])(
    'turns %j into NaN',
    (param) => {
      expect(parseSnippetId(param)).toBeNaN()
    },
  )
})
