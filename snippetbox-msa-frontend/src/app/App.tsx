import { BrowserRouter } from 'react-router-dom'
import { AppProviders } from './AppProviders'
import { AppRoutes } from './AppRoutes'

export function App() {
  return (
    <BrowserRouter useTransitions={false}>
      <AppProviders>
        <AppRoutes />
      </AppProviders>
    </BrowserRouter>
  )
}
