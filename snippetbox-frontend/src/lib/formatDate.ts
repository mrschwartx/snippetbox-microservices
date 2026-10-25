const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

const pad = (value: number): string => String(value).padStart(2, '0')

export const formatDateTime = (iso: string): string => {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''

  const day = pad(date.getUTCDate())
  const month = MONTHS[date.getUTCMonth()]
  const year = date.getUTCFullYear()
  const time = `${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}`
  return `${day} ${month} ${year} at ${time}`
}
