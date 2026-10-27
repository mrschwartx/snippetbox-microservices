export const snippetKeys = {
  all: ['snippets'] as const,
  latest: () => [...snippetKeys.all, 'latest'] as const,
  detail: (id: number) => [...snippetKeys.all, 'detail', id] as const,
}
