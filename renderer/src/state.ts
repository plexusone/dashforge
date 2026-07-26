type Listener = (key: string, value: unknown) => void

export class PageState {
  private values: Record<string, unknown> = {}
  private listeners = new Map<string, Set<Listener>>()

  get(path: string): unknown {
    const segments = path.split('.')
    let current: unknown = this.values
    for (const seg of segments) {
      if (current === null || current === undefined) return undefined
      if (typeof current === 'object') {
        current = (current as Record<string, unknown>)[seg]
      } else {
        return undefined
      }
    }
    return current
  }

  set(path: string, value: unknown): void {
    const segments = path.split('.')
    if (segments.length === 1) {
      this.values[path] = value
      this.notify(path, value)
      return
    }

    let current: Record<string, unknown> = this.values
    for (let i = 0; i < segments.length - 1; i++) {
      const seg = segments[i]
      if (!(seg in current) || typeof current[seg] !== 'object' || current[seg] === null) {
        current[seg] = {}
      }
      current = current[seg] as Record<string, unknown>
    }
    current[segments[segments.length - 1]] = value
    this.notify(path, value)
  }

  toggle(path: string): void {
    const current = this.get(path)
    if (typeof current !== 'boolean') {
      throw new Error(`Cannot toggle non-boolean value at "${path}": got ${typeof current}`)
    }
    this.set(path, !current)
  }

  subscribe(pathPrefix: string, fn: Listener): () => void {
    let listeners = this.listeners.get(pathPrefix)
    if (!listeners) {
      listeners = new Set()
      this.listeners.set(pathPrefix, listeners)
    }
    listeners.add(fn)
    return () => {
      listeners!.delete(fn)
      if (listeners!.size === 0) {
        this.listeners.delete(pathPrefix)
      }
    }
  }

  snapshot(): Record<string, unknown> {
    return JSON.parse(JSON.stringify(this.values))
  }

  load(initial: Record<string, unknown>): void {
    this.values = JSON.parse(JSON.stringify(initial))
  }

  private notify(path: string, value: unknown): void {
    for (const [prefix, listeners] of this.listeners) {
      if (path === prefix || path.startsWith(prefix + '.') || prefix === '*') {
        for (const fn of listeners) {
          fn(path, value)
        }
      }
    }
  }
}
