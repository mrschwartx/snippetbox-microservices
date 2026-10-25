import { useMutation, useQueryClient } from '@tanstack/react-query'
import { logout } from '../api/authApi'
import { authKeys } from '../queryKeys'

export function useLogout() {
  const queryClient = useQueryClient()
  const mutation = useMutation({ mutationFn: logout })

  const signOut = (onSignedOut?: () => void) =>
    mutation.mutate(undefined, {
      onSuccess: () => {
        onSignedOut?.()
        queryClient.setQueryData(authKeys.currentUser, null)
      },
    })

  return { signOut, isPending: mutation.isPending }
}
