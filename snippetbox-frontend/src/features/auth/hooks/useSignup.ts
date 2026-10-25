import { useMutation } from '@tanstack/react-query'
import { signup } from '../api/authApi'

export function useSignup() {
  return useMutation({ mutationFn: signup })
}
