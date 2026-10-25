import { useCallback, useMemo, useState, type ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { FlashContext } from './FlashContext'
import { advanceFlash, queueFlash, visibleFlashMessage, type PendingFlash } from './flashState'

interface FlashProviderProps {
  children: ReactNode
}

export function FlashProvider({ children }: FlashProviderProps) {
  const { key } = useLocation()
  const [flash, setFlash] = useState<PendingFlash | null>(null)

  const advanced = advanceFlash(flash, key)
  if (advanced !== flash) setFlash(advanced)

  const push = useCallback((message: string) => setFlash(queueFlash(message, key)), [key])

  const message = visibleFlashMessage(advanced, key)
  const value = useMemo(() => ({ message, push }), [message, push])

  return <FlashContext.Provider value={value}>{children}</FlashContext.Provider>
}
