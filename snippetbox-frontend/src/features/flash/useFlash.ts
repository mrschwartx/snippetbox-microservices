import { useContext } from 'react'
import { FlashContext, type FlashContextValue } from './FlashContext'

export function useFlash(): FlashContextValue {
  const context = useContext(FlashContext)
  if (!context) throw new Error('useFlash must be used inside FlashProvider')
  return context
}
