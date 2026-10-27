import { QueryClientProvider, type QueryClient } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { FlashProvider } from '@/features/flash'
import { createQueryClient } from './queryClient'

interface AppProvidersProps {
  children: ReactNode
  queryClient?: QueryClient
}

export function AppProviders({ children, queryClient }: AppProvidersProps) {
  const [client] = useState<QueryClient>(() => queryClient ?? createQueryClient())

  return (
    <QueryClientProvider client={client}>
      <FlashProvider>{children}</FlashProvider>
    </QueryClientProvider>
  )
}
