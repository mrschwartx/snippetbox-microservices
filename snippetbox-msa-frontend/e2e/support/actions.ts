import type { Page } from '@playwright/test'
import { DEFAULT_USER, type TestUser } from '../../src/test/fakeBackend'

export const signUp = async (page: Page, user: TestUser = DEFAULT_USER) => {
  await page.goto('/user/signup')
  await page.getByLabel('Name:').fill(user.name)
  await page.getByLabel('Email:').fill(user.email)
  await page.getByLabel('Password:').fill(user.password)
  await page.getByRole('button', { name: 'Signup' }).click()
}

export const logIn = async (page: Page, user: TestUser = DEFAULT_USER) => {
  await page.goto('/user/login')
  await page.getByLabel('Email:').fill(user.email)
  await page.getByLabel('Password:').fill(user.password)
  await page.getByRole('button', { name: 'Login' }).click()
}

export const publishSnippet = async (
  page: Page,
  { title, content, expiry }: { title: string; content: string; expiry?: string },
) => {
  await page.getByLabel('Title:').fill(title)
  await page.getByLabel('Content:').fill(content)
  if (expiry) await page.getByLabel(expiry).check()
  await page.getByRole('button', { name: 'Publish snippet' }).click()
}
