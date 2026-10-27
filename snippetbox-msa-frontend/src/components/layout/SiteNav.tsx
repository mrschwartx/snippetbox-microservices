import { flushSync } from 'react-dom'
import { NavLink, useNavigate } from 'react-router-dom'
import { useCurrentUser, useLogout } from '@/features/auth'
import { useFlash } from '@/features/flash'
import { linkButtonClass } from '../ui/classNames'
import { CONTAINER_CLASS } from './containerClass'

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'cursor-default text-ink hover:text-ink hover:no-underline' : undefined

export function SiteNav() {
  const navigate = useNavigate()
  const flash = useFlash()
  const { data: user, isPending } = useCurrentUser()
  const { signOut } = useLogout()

  const handleLogout = () => {
    signOut(() => {
      flushSync(() => {
        flash.push("You've been logged out successfully!")
        navigate('/')
      })
    })
  }

  return (
    <nav className="border-b border-line bg-surface-muted text-ink-muted">
      <div className={`${CONTAINER_CLASS} flex min-h-[60px] items-center justify-between`}>
        <div className="flex gap-6">
          <NavLink to="/" end className={navLinkClass}>
            Home
          </NavLink>
          {user && (
            <NavLink to="/snippet/create" className={navLinkClass}>
              Create snippet
            </NavLink>
          )}
        </div>
        <div className="flex gap-6">
          {!isPending && user && (
            <button type="button" className={linkButtonClass} onClick={handleLogout}>
              Logout
            </button>
          )}
          {!isPending && !user && (
            <>
              <NavLink to="/user/signup" className={navLinkClass}>
                Signup
              </NavLink>
              <NavLink to="/user/login" className={navLinkClass}>
                Login
              </NavLink>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
