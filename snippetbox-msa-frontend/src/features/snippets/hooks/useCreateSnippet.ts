import { useMutation, useQueryClient } from '@tanstack/react-query'
import { authKeys } from '@/features/auth'
import { UnauthenticatedError } from '@/lib/errors'
import { createSnippet } from '../api/snippetsApi'
import { snippetKeys } from '../queryKeys'

export function useCreateSnippet() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createSnippet,
    onSuccess: (snippet) => {
      queryClient.setQueryData(snippetKeys.detail(snippet.id), snippet)
      void queryClient.invalidateQueries({ queryKey: snippetKeys.latest() })
    },
    onError: (error) => {
      if (error instanceof UnauthenticatedError) {
        void queryClient.invalidateQueries({ queryKey: authKeys.currentUser })
      }
    },
  })
}
