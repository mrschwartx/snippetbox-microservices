export const parseSnippetId = (param: string | undefined): number =>
  /^\d+$/.test(param ?? '') ? Number(param) : Number.NaN
