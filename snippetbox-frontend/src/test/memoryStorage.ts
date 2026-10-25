export const createMemoryStorage = () => {
  const data = new Map<string, string>()

  return {
    data,
    getItem: (key: string) => data.get(key) ?? null,
    setItem: (key: string, value: string) => void data.set(key, value),
    removeItem: (key: string) => void data.delete(key),
  }
}

export const createBlockedStorage = () => {
  const blocked = () => {
    throw new Error('storage is blocked')
  }

  return { getItem: blocked, setItem: blocked, removeItem: blocked }
}
