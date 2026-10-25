import { expect, test } from './fixtures'
import { logIn } from './support/actions'

test.describe('navigation and layout', () => {
  test('the brand link returns to the home page', async ({ page }) => {
    await page.goto('/user/login')

    await page.getByRole('link', { name: 'Snippetbox' }).click()

    await expect(page).toHaveURL('/')
  })

  test('a visitor sees Home, Signup and Login', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Signup' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Login' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Create snippet' })).toHaveCount(0)
  })

  test('a signed in user sees Create snippet and Logout', async ({ page, backend }) => {
    backend.seedUser()
    await logIn(page)

    await expect(page.getByRole('link', { name: 'Create snippet' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Logout' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Login' })).toHaveCount(0)
  })

  test('the link of the current page is marked as current', async ({ page }) => {
    await page.goto('/user/signup')

    await expect(page.getByRole('link', { name: 'Signup' })).toHaveAttribute('aria-current', 'page')
    await expect(page.getByRole('link', { name: 'Home' })).not.toHaveAttribute('aria-current')
  })

  test('an unknown route shows Not Found inside the layout', async ({ page }) => {
    await page.goto('/no/such/page')

    await expect(page.getByText('Not Found')).toBeVisible()
    await expect(page.getByRole('link', { name: 'Home' })).toBeVisible()
    await expect(page.getByRole('contentinfo')).toContainText('Powered by')
  })

  test('the home page reports an API outage instead of staying blank', async ({ page }) => {
    await page.route('**/api/v1/snippets', (route) => route.fulfill({ status: 503, json: {} }))

    await page.goto('/')

    await expect(page.getByRole('alert')).toHaveText('Something went wrong. Please try again.')
  })
})

test.describe('appearance', () => {
  test('follows the light colour scheme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' })
    await page.goto('/')

    await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(241, 243, 246)')
  })

  test('follows the dark colour scheme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' })
    await page.goto('/')

    await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(20, 24, 29)')
  })

  test('does not scroll sideways on a phone, even with long snippet content', async ({
    page,
    backend,
  }) => {
    backend.seedSnippet({ title: 'Wide', content: 'x'.repeat(300) })
    await page.setViewportSize({ width: 375, height: 667 })

    await page.goto('/snippet/view/1')
    await expect(page.getByRole('article')).toBeVisible()

    const overflows = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
    )
    expect(overflows).toBe(false)
  })
})
