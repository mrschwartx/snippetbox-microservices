import { renderHook } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { FlashProvider } from './FlashProvider'
import { useFlash } from './useFlash'

describe('useFlash', () => {
  it('exposes no message and a push function inside the provider', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <MemoryRouter>
        <FlashProvider>{children}</FlashProvider>
      </MemoryRouter>
    )

    const { result } = renderHook(() => useFlash(), { wrapper })

    expect(result.current.message).toBeNull()
    expect(result.current.push).toBeTypeOf('function')
  })

  it('refuses to be used outside a FlashProvider', () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)

    expect(() => renderHook(() => useFlash())).toThrow('useFlash must be used inside FlashProvider')

    vi.restoreAllMocks()
  })
})
