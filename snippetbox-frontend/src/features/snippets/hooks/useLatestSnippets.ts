import { useQuery } from '@tanstack/react-query'
import { getLatestSnippets } from '../api/snippetsApi'
import { snippetKeys } from '../queryKeys'

export function useLatestSnippets() {
  return useQuery({
    queryKey: snippetKeys.latest(),
    queryFn: getLatestSnippets,
  })
}
