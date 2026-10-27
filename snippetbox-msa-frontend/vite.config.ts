import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath } from 'node:url'
import { configDefaults, defineConfig } from 'vitest/config'

const TEST_API_BASE_URL = 'http://snippetbox.test/api/v1'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api/v1/auth': {
        target: process.env.VITE_USER_PROXY_TARGET ?? 'http://localhost:8081',
        changeOrigin: true,
      },
      '/api/v1/snippets': {
        target: process.env.VITE_SNIPPET_PROXY_TARGET ?? 'http://localhost:8082',
        changeOrigin: true,
      },
    },
  },
  preview: { port: 3000, host: true },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
    env: { VITE_API_BASE_URL: TEST_API_BASE_URL },
    coverage: {
      provider: 'v8',
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/test/**',
        'src/**/*.test.{ts,tsx}',
        'src/**/index.ts',
        'src/**/types.ts',
        'src/main.tsx',
        'src/vite-env.d.ts',
      ],
      thresholds: { statements: 95, branches: 90, functions: 95, lines: 95 },
      reporter: ['text', 'html'],
    },
    projects: [
      {
        extends: true,
        test: {
          name: 'unit',
          include: ['src/**/*.test.{ts,tsx}'],
          exclude: [...configDefaults.exclude, 'src/**/*.integration.test.{ts,tsx}'],
        },
      },
      {
        extends: true,
        test: {
          name: 'integration',
          include: ['src/**/*.integration.test.{ts,tsx}'],
        },
      },
    ],
  },
})
