import { Navigate, Outlet } from 'react-router-dom'
import { useCurrentUser } from '../hooks/useCurrentUser'

export function RequireAuth() {
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null
  if (!user) return <Navigate to="/user/login" replace />
  return <Outlet />
}
