import { test as base, expect } from '@playwright/test'
import { createFakeBackend, type FakeBackend } from '../src/test/fakeBackend'

export const FIXED_NOW = new Date('2026-10-15T08:00:00.000Z')

const API_PATTERN = '**/api/v1/**'
const API_PATH_PREFIX = '/api/v1'

interface Fixtures {
  backend: FakeBackend
}

export const test = base.extend<Fixtures>({
  backend: async ({}, use) => {
    await use(createFakeBackend({ now: () => FIXED_NOW }))
  },

  page: async ({ page, backend, baseURL }, use) => {
    const appOrigin = new URL(baseURL ?? 'http://localhost:3000').origin

    await page.clock.setFixedTime(FIXED_NOW)

    await page.route(
      (url) => url.origin !== appOrigin,
      (route) => route.abort(),
    )

    await page.route(API_PATTERN, async (route) => {
      const request = route.request()
      const postData = request.postData()
      const authorization = request.headers().authorization

      const { status, body } = backend.handle({
        method: request.method(),
        path: new URL(request.url()).pathname.slice(API_PATH_PREFIX.length),
        token: authorization?.replace(/^Bearer /, '') ?? null,
        body: postData ? JSON.parse(postData) : undefined,
      })

      await route.fulfill(
        body === undefined
          ? { status }
          : { status, contentType: 'application/json', body: JSON.stringify(body) },
      )
    })

    await use(page)
  },
})

export { expect }
