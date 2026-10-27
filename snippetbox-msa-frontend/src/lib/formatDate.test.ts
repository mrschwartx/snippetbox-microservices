import { describe, expect, it } from 'vitest'
import { formatDateTime } from './formatDate'

describe('formatDateTime', () => {
  it.each([
    ['2026-10-15T08:05:00Z', '15 Oct 2026 at 08:05'],
    ['2026-01-01T00:00:00Z', '01 Jan 2026 at 00:00'],
    ['2026-12-31T23:59:59Z', '31 Dec 2026 at 23:59'],
    ['2026-02-09T09:09:00.000Z', '09 Feb 2026 at 09:09'],
  ])('formats %s as %s', (iso, expected) => {
    expect(formatDateTime(iso)).toBe(expected)
  })

  it('always renders in UTC, whatever offset the input carries', () => {
    expect(formatDateTime('2026-10-15T10:00:00+02:00')).toBe('15 Oct 2026 at 08:00')
    expect(formatDateTime('2026-10-15T01:00:00-07:00')).toBe('15 Oct 2026 at 08:00')
  })

  it.each(['', 'not a date', '2026-13-45'])('returns an empty string for %j', (iso) => {
    expect(formatDateTime(iso)).toBe('')
  })
})
