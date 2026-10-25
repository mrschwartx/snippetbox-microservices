import { useNavigate } from 'react-router-dom'
import { SignupForm } from '@/features/auth'
import { useFlash } from '@/features/flash'

export function SignupPage() {
  const navigate = useNavigate()
  const flash = useFlash()

  const handleSuccess = () => {
    flash.push('Your signup was successful. Please log in.')
    navigate('/user/login')
  }

  return <SignupForm onSuccess={handleSuccess} />
}
