import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactElement } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { AppProviders } from '@/app/AppProviders'
import { AppRoutes } from '@/app/AppRoutes'
import { createQueryClient } from '@/app/queryClient'

interface RenderOptions {
  route?: string
}

export const renderWithProviders = (ui: ReactElement, { route = '/' }: RenderOptions = {}) => {
  const user = userEvent.setup()
  const queryClient = createQueryClient()
  const result = render(
    <MemoryRouter initialEntries={[route]} useTransitions={false}>
      <AppProviders queryClient={queryClient}>{ui}</AppProviders>
    </MemoryRouter>,
  )
  return { user, queryClient, ...result }
}

export const renderApp = (route = '/') => renderWithProviders(<AppRoutes />, { route })
