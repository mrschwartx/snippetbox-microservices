# Snippetbox Frontend

React 19 + TypeScript + Vite + Tailwind CSS 4, with TanStack Query and React Router.

## Getting started

```sh
npm install
cp .env.example .env
npm run dev
```

`VITE_API_BASE_URL` is the API root used by the browser (default `/api/v1`).
`VITE_USER_PROXY_TARGET` is where the dev server proxies `/api/v1/auth` (default `http://localhost:8080`).
`VITE_SNIPPET_PROXY_TARGET` is where it proxies `/api/v1/snippets` (default `http://localhost:8082`).

## Scripts

| Script                     | What it does                              |
| -------------------------- | ----------------------------------------- |
| `npm run dev`              | Start the dev server on port 3000         |
| `npm run build`            | Typecheck and build for production        |
| `npm run lint`             | ESLint                                    |
| `npm run typecheck`        | `tsc -b` for app, config and e2e projects |
| `npm run format`           | Prettier (write)                          |
| `npm test`                 | Unit and integration tests (Vitest)       |
| `npm run test:unit`        | Unit tests only                           |
| `npm run test:integration` | Integration tests only                    |
| `npm run test:coverage`    | Vitest with coverage report               |
| `npm run test:e2e`         | End-to-end tests (Playwright)             |

End-to-end tests need a browser once: `npx playwright install chromium`.
Set `PLAYWRIGHT_CHROMIUM_PATH` to use an existing Chromium binary instead.

## Structure

```
src/
  app/            App, providers, routes, query client
  components/
    layout/       Layout, header, nav, footer
    ui/           Generic form and feedback components
  features/
    auth/         api, components, hooks, lib (validation), types
    snippets/     api, components, hooks, lib (validation), types
    flash/        One-shot flash messages across navigation
  lib/            API client, errors, session store, validator, formatting
  pages/          One component per route, composed from features
  styles/         Global CSS and theme tokens
  test/           Test setup, MSW server, fake backend, render helpers
e2e/              Playwright specs, fixtures and helpers
```

Rules of thumb:

- A feature owns its api, hooks, components and validation. Other code imports it through its `index.ts`.
- `lib/` and `components/ui/` never import from `features/`.
- Pages stay thin: they wire a feature to routing and flash messages.

## Testing

- **Unit** (`*.test.ts(x)`, next to the code): pure functions, API layer, presentational components.
- **Integration** (`*.integration.test.tsx`): real pages, router and React Query, with the network
  replaced by MSW backed by `src/test/fakeBackend.ts`.
- **End-to-end** (`e2e/*.spec.ts`): a real browser against the dev server. The API is mocked at the
  network level with the same fake backend, time is frozen, and external requests are blocked.
