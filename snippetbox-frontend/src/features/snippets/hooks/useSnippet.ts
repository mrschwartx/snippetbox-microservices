import { useQuery } from '@tanstack/react-query'
import { getSnippet } from '../api/snippetsApi'
import { snippetKeys } from '../queryKeys'

export function useSnippet(id: number) {
  return useQuery({
    queryKey: snippetKeys.detail(id),
    queryFn: () => getSnippet(id),
  })
}
