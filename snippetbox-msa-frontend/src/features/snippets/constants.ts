export const TITLE_MAX_LENGTH = 100
export const DEFAULT_EXPIRES_DAYS = 365

export const EXPIRY_OPTIONS = [
  { days: 365, label: 'One Year' },
  { days: 7, label: 'One Week' },
  { days: 1, label: 'One Day' },
] as const

export const PERMITTED_EXPIRES: readonly number[] = EXPIRY_OPTIONS.map((option) => option.days)
