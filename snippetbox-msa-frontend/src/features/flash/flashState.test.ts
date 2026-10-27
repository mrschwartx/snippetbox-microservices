import { describe, expect, it } from 'vitest'
import { advanceFlash, queueFlash, visibleFlashMessage } from './flashState'

describe('queueFlash', () => {
  it('remembers the page it was queued from and is not yet shown anywhere', () => {
    expect(queueFlash('Saved', 'page-a')).toEqual({
      message: 'Saved',
      queuedFrom: 'page-a',
      shownOn: null,
    })
  })
})

describe('advanceFlash', () => {
  it('has nothing to advance when no flash is pending', () => {
    expect(advanceFlash(null, 'page-a')).toBeNull()
  })

  it('keeps a queued flash waiting while the user is still on the page that queued it', () => {
    const flash = queueFlash('Saved', 'page-a')

    expect(advanceFlash(flash, 'page-a')).toBe(flash)
  })

  it('pins the flash to the first page visited after it was queued', () => {
    const flash = queueFlash('Saved', 'page-a')

    expect(advanceFlash(flash, 'page-b')).toEqual({ ...flash, shownOn: 'page-b' })
  })

  it('keeps the flash while the user stays on the page it was pinned to', () => {
    const pinned = { message: 'Saved', queuedFrom: 'page-a', shownOn: 'page-b' }

    expect(advanceFlash(pinned, 'page-b')).toBe(pinned)
  })

  it('drops the flash as soon as the user navigates again', () => {
    const pinned = { message: 'Saved', queuedFrom: 'page-a', shownOn: 'page-b' }

    expect(advanceFlash(pinned, 'page-c')).toBeNull()
  })

  it('drops the flash even when the user returns to the page that queued it', () => {
    const pinned = { message: 'Saved', queuedFrom: 'page-a', shownOn: 'page-b' }

    expect(advanceFlash(pinned, 'page-a')).toBeNull()
  })
})

describe('visibleFlashMessage', () => {
  it('shows nothing without a flash', () => {
    expect(visibleFlashMessage(null, 'page-a')).toBeNull()
  })

  it('does not show a flash on the page that queued it', () => {
    expect(visibleFlashMessage(queueFlash('Saved', 'page-a'), 'page-a')).toBeNull()
  })

  it('shows a flash only on the page it was pinned to', () => {
    const pinned = { message: 'Saved', queuedFrom: 'page-a', shownOn: 'page-b' }

    expect(visibleFlashMessage(pinned, 'page-b')).toBe('Saved')
    expect(visibleFlashMessage(pinned, 'page-c')).toBeNull()
  })
})
