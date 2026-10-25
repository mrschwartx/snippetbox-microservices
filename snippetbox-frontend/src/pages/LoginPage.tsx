import { useNavigate } from 'react-router-dom'
import { LoginForm } from '@/features/auth'

export function LoginPage() {
  const navigate = useNavigate()

  return <LoginForm onSuccess={() => navigate('/snippet/create')} />
}
