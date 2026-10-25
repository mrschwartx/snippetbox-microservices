import { QueryClient } from '@tanstack/react-query'

export const createQueryClient = (): QueryClient =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: 15_000, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  })
