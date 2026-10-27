import { Outlet } from 'react-router-dom'
import { FlashMessage } from '@/features/flash'
import { CONTAINER_CLASS } from './containerClass'
import { SiteFooter } from './SiteFooter'
import { SiteHeader } from './SiteHeader'
import { SiteNav } from './SiteNav'

export function Layout() {
  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      <SiteNav />
      <main className={`${CONTAINER_CLASS} my-[54px] flex-1`}>
        <FlashMessage />
        <Outlet />
      </main>
      <SiteFooter />
    </div>
  )
}
