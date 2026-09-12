// Plain Node (our default 'node' test environment) has no global
// `localStorage` at all, so without this, lib/tokens.ts would silently
// degrade to "storage unavailable" (it wraps every access in try/catch) and
// refresh-token persistence would be untestable. A jsdom-environment test
// file (`// @vitest-environment jsdom`) already gets a real one from jsdom
// as a getter-only property on Window, which we must NOT clobber — hence
// the probe below rather than an unconditional assignment.
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

function localStorageWorks(): boolean {
  try {
    globalThis.localStorage.setItem('__probe__', '1')
    globalThis.localStorage.removeItem('__probe__')
    return true
  } catch {
    return false
  }
}

if (!localStorageWorks()) {
  Object.defineProperty(globalThis, 'localStorage', {
    value: new MemoryStorage(),
    writable: true,
    configurable: true,
  })
}
