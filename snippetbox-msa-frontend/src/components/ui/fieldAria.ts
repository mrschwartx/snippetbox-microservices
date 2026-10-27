export const getFieldAriaProps = (id: string, error?: string) => ({
  'aria-invalid': error ? true : undefined,
  'aria-describedby': error ? `${id}-error` : undefined,
})
