import { useQuery } from '@tanstack/react-query'
import { getCurrentUser } from '../api/authApi'
import { authKeys } from '../queryKeys'

export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.currentUser,
    queryFn: getCurrentUser,
    staleTime: Infinity,
  })
}
