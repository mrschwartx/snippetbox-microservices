export interface PendingFlash {
  message: string
  queuedFrom: string
  shownOn: string | null
}

export const queueFlash = (message: string, locationKey: string): PendingFlash => ({
  message,
  queuedFrom: locationKey,
  shownOn: null,
})

export const advanceFlash = (
  flash: PendingFlash | null,
  locationKey: string,
): PendingFlash | null => {
  if (!flash) return null

  if (flash.shownOn === null) {
    return locationKey === flash.queuedFrom ? flash : { ...flash, shownOn: locationKey }
  }

  return locationKey === flash.shownOn ? flash : null
}

export const visibleFlashMessage = (
  flash: PendingFlash | null,
  locationKey: string,
): string | null => (flash && flash.shownOn === locationKey ? flash.message : null)
