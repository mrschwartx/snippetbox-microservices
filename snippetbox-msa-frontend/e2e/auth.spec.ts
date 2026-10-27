import { DEFAULT_USER } from '../src/test/fakeBackend'
import { FIXED_NOW, expect, test } from './fixtures'
import { logIn, signUp } from './support/actions'

const HOUR_MS = 60 * 60 * 1000

test.describe('authentication', () => {
  test('a visitor signs up, logs in and logs out', async ({ page }) => {
    await signUp(page)

    await expect(page).toHaveURL('/user/login')
    await expect(page.getByRole('status')).toHaveText('Your signup was successful. Please log in.')

    await page.getByLabel('Email:').fill(DEFAULT_USER.email)
    await page.getByLabel('Password:').fill(DEFAULT_USER.password)
    await page.getByRole('button', { name: 'Login' }).click()

    await expect(page).toHaveURL('/snippet/create')
    await expect(page.getByRole('button', { name: 'Logout' })).toBeVisible()

    await page.getByRole('button', { name: 'Logout' }).click()

    await expect(page).toHaveURL('/')
    await expect(page.getByRole('status')).toHaveText("You've been logged out successfully!")
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible()
  })

  test('signup flags every blank field without calling the API', async ({ page, backend }) => {
    await page.goto('/user/signup')

    await page.getByRole('button', { name: 'Signup' }).click()

    await expect(page.getByText('This field cannot be blank')).toHaveCount(3)
    expect(backend.requests()).toEqual([])
  })

  test('signup rejects a short password', async ({ page }) => {
    await page.goto('/user/signup')
    await page.getByLabel('Name:').fill('Alice')
    await page.getByLabel('Email:').fill('alice@example.com')
    await page.getByLabel('Password:').fill('short')

    await page.getByRole('button', { name: 'Signup' }).click()

    await expect(page.getByText('This field must be at least 8 characters long')).toBeVisible()
    await expect(page.getByLabel('Password:')).toHaveAttribute('aria-invalid', 'true')
  })

  test('signup reports an email that is already registered', async ({ page, backend }) => {
    backend.seedUser()

    await signUp(page, { ...DEFAULT_USER, name: 'Eve' })

    await expect(page.getByText('Email address is already in use')).toBeVisible()
    await expect(page).toHaveURL('/user/signup')
  })

  test('login rejects wrong credentials', async ({ page, backend }) => {
    backend.seedUser()

    await logIn(page, { ...DEFAULT_USER, password: 'wrong-password' })

    await expect(page.getByText('Email or password is incorrect')).toBeVisible()
    await expect(page).toHaveURL('/user/login')
  })

  test('login can be completed with the keyboard alone', async ({ page, backend }) => {
    backend.seedUser()
    await page.goto('/user/login')

    await page.getByLabel('Email:').focus()
    await page.keyboard.type(DEFAULT_USER.email)
    await page.keyboard.press('Tab')
    await page.keyboard.type(DEFAULT_USER.password)
    await page.keyboard.press('Enter')

    await expect(page).toHaveURL('/snippet/create')
  })

  test('an anonymous visitor is redirected away from the create page', async ({ page }) => {
    await page.goto('/snippet/create')

    await expect(page).toHaveURL('/user/login')
    await expect(page.getByLabel('Title:')).toHaveCount(0)
  })

  test('the session survives a page reload', async ({ page, backend }) => {
    backend.seedUser()
    await logIn(page)
    await expect(page).toHaveURL('/snippet/create')

    await page.reload()

    await expect(page.getByRole('button', { name: 'Logout' })).toBeVisible()
    await expect(page.getByLabel('Title:')).toBeVisible()
  })

  test('a session older than its lifetime no longer signs the user in', async ({
    page,
    backend,
  }) => {
    backend.seedUser()
    await logIn(page)
    await expect(page).toHaveURL('/snippet/create')

    await page.clock.setFixedTime(new Date(FIXED_NOW.getTime() + 13 * HOUR_MS))
    await page.reload()

    await expect(page).toHaveURL('/user/login')
    await expect(page.getByRole('link', { name: 'Signup' })).toBeVisible()
  })
})
