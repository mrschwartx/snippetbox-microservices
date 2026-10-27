import { createContext } from 'react'

export interface FlashContextValue {
  message: string | null
  push: (message: string) => void
}

export const FlashContext = createContext<FlashContextValue | null>(null)
