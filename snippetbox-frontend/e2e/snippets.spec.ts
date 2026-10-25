import { expect, test } from './fixtures'
import { logIn, publishSnippet } from './support/actions'

test.describe('browsing snippets', () => {
  test('the home page invites the first post when there are no snippets', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('heading', { name: 'Latest Snippets' })).toBeVisible()
    await expect(page.getByText("There's nothing to see here... yet!")).toBeVisible()
  })

  test('the home page lists snippets newest first and opens one', async ({ page, backend }) => {
    backend.seedSnippet({ title: 'Older snippet', content: 'old' })
    backend.seedSnippet({ title: 'Newer snippet', content: 'new content', days: 1 })
    await page.goto('/')

    const titles = page.getByRole('row').getByRole('link')
    await expect(titles).toHaveText(['Newer snippet', 'Older snippet'])

    await titles.first().click()

    await expect(page).toHaveURL('/snippet/view/2')
    const article = page.getByRole('article')
    await expect(article).toContainText('Newer snippet')
    await expect(article).toContainText('new content')
    await expect(article).toContainText('#2')
    await expect(article).toContainText('Created: 15 Oct 2026 at 08:00')
    await expect(article).toContainText('Expires: 16 Oct 2026 at 08:00')
  })

  test('an unknown snippet shows Not Found', async ({ page }) => {
    await page.goto('/snippet/view/99')

    await expect(page.getByText('Not Found')).toBeVisible()
  })

  test('a malformed snippet id shows Not Found without asking the API', async ({
    page,
    backend,
  }) => {
    await page.goto('/snippet/view/abc')

    await expect(page.getByText('Not Found')).toBeVisible()
    expect(backend.requests()).toEqual([])
  })

  test('snippet content is shown as text, never executed', async ({ page, backend }) => {
    backend.seedSnippet({ title: 'Markup', content: '<img src=x onerror="window.pwned=1">' })
    await page.goto('/snippet/view/1')

    await expect(page.getByRole('article')).toContainText('<img src=x onerror="window.pwned=1">')
    expect(await page.evaluate(() => 'pwned' in window)).toBe(false)
  })
})

test.describe('publishing snippets', () => {
  test.beforeEach(async ({ page, backend }) => {
    backend.seedUser()
    await logIn(page)
    await expect(page).toHaveURL('/snippet/create')
  })

  test('a signed in user publishes a snippet and sees it', async ({ page }) => {
    await publishSnippet(page, { title: 'Hello world', content: 'fmt.Println("hi")' })

    await expect(page).toHaveURL('/snippet/view/1')
    await expect(page.getByRole('status')).toHaveText('Snippet successfully created!')
    await expect(page.getByRole('article')).toContainText('Hello world')
    await expect(page.getByRole('article')).toContainText('fmt.Println("hi")')
  })

  const expiries = [
    ['One Day', 'Expires: 16 Oct 2026 at 08:00'],
    ['One Week', 'Expires: 22 Oct 2026 at 08:00'],
    ['One Year', 'Expires: 15 Oct 2027 at 08:00'],
  ] as const

  for (const [expiry, expected] of expiries) {
    test(`choosing "${expiry}" sets the expiry date`, async ({ page }) => {
      await publishSnippet(page, { title: 'T', content: 'C', expiry })

      await expect(page.getByRole('article')).toContainText(expected)
    })
  }

  test('the flash message disappears on the next navigation and the snippet is listed', async ({
    page,
  }) => {
    await publishSnippet(page, { title: 'Listed snippet', content: 'C' })
    await expect(page.getByRole('status')).toBeVisible()

    await page.getByRole('link', { name: 'Home' }).click()

    await expect(page.getByRole('status')).toHaveCount(0)
    await expect(page.getByRole('link', { name: 'Listed snippet' })).toBeVisible()
  })

  test('blank fields are reported and the form is kept', async ({ page, backend }) => {
    backend.clearRequests()

    await page.getByRole('button', { name: 'Publish snippet' }).click()

    await expect(page.getByText('This field cannot be blank')).toHaveCount(2)
    await expect(page.getByLabel('One Year')).toBeChecked()
    expect(backend.requests()).toEqual([])
  })

  test('a title longer than 100 characters is rejected', async ({ page }) => {
    await publishSnippet(page, { title: 'a'.repeat(101), content: 'C' })

    await expect(page.getByText('This field cannot be more than 100 characters long')).toBeVisible()
  })

  test('a server failure keeps what was typed', async ({ page }) => {
    await page.route('**/api/v1/snippets', (route) =>
      route.request().method() === 'POST'
        ? route.fulfill({ status: 500, json: {} })
        : route.fallback(),
    )

    await publishSnippet(page, { title: 'Keep me', content: 'Keep this too' })

    await expect(page.getByRole('alert')).toHaveText('Something went wrong. Please try again.')
    await expect(page.getByLabel('Title:')).toHaveValue('Keep me')
    await expect(page.getByLabel('Content:')).toHaveValue('Keep this too')
  })

  test('a session revoked by the server sends the user back to login', async ({
    page,
    backend,
  }) => {
    backend.reset()

    await publishSnippet(page, { title: 'T', content: 'C' })

    await expect(page).toHaveURL('/user/login')
  })
})
