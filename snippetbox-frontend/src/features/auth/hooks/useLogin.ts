import { useMutation, useQueryClient } from '@tanstack/react-query'
import { login } from '../api/authApi'
import { authKeys } from '../queryKeys'

export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: login,
    onSuccess: (user) => {
      queryClient.setQueryData(authKeys.currentUser, user)
    },
  })
}
