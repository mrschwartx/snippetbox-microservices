import { useFlash } from './useFlash'

export function FlashMessage() {
  const { message } = useFlash()
  if (!message) return null

  return (
    <div role="status" className="mb-9 bg-flash p-[1rem] text-center font-bold text-white">
      {message}
    </div>
  )
}
