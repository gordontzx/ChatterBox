// Node has no global `localStorage` (its experimental Web Storage API isn't
// enabled here, and isn't worth depending on for tests anyway — it can
// persist to disk). lib/tokens.ts wraps every access in try/catch, so
// without this it would silently degrade to "storage unavailable"; this
// polyfill makes refresh-token persistence actually testable.
class MemoryStorage implements Storage {
  private store = new Map<string, string>()

  get length(): number {
    return this.store.size
  }

  getItem(key: string): string | null {
    return this.store.has(key) ? (this.store.get(key) ?? null) : null
  }

  setItem(key: string, value: string): void {
    this.store.set(key, value)
  }

  removeItem(key: string): void {
    this.store.delete(key)
  }

  clear(): void {
    this.store.clear()
  }

  key(index: number): string | null {
    return Array.from(this.store.keys())[index] ?? null
  }
}

globalThis.localStorage = new MemoryStorage()
