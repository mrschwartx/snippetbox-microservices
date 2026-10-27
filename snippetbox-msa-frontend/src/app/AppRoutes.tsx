import { Route, Routes } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { RequireAuth } from '@/features/auth'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { SignupPage } from '@/pages/SignupPage'
import { SnippetCreatePage } from '@/pages/SnippetCreatePage'
import { SnippetViewPage } from '@/pages/SnippetViewPage'

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="snippet/view/:id" element={<SnippetViewPage />} />
        <Route path="user/signup" element={<SignupPage />} />
        <Route path="user/login" element={<LoginPage />} />
        <Route element={<RequireAuth />}>
          <Route path="snippet/create" element={<SnippetCreatePage />} />
        </Route>
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}
