import { describe, it, expect, vi } from 'vitest'
import { PageState } from './state'

describe('PageState', () => {
  it('gets and sets simple values', () => {
    const s = new PageState()
    s.set('name', 'Alice')
    expect(s.get('name')).toBe('Alice')
  })

  it('gets and sets nested paths', () => {
    const s = new PageState()
    s.set('filters.month', 'July')
    expect(s.get('filters.month')).toBe('July')
    expect(s.get('filters')).toEqual({ month: 'July' })
  })

  it('sets deeply nested paths creating intermediates', () => {
    const s = new PageState()
    s.set('a.b.c', 42)
    expect(s.get('a.b.c')).toBe(42)
    expect(s.get('a.b')).toEqual({ c: 42 })
  })

  it('returns undefined for missing paths', () => {
    const s = new PageState()
    expect(s.get('missing')).toBeUndefined()
    expect(s.get('a.b.c')).toBeUndefined()
  })

  it('toggles boolean values', () => {
    const s = new PageState()
    s.set('active', true)
    s.toggle('active')
    expect(s.get('active')).toBe(false)
    s.toggle('active')
    expect(s.get('active')).toBe(true)
  })

  it('throws on toggle of non-boolean', () => {
    const s = new PageState()
    s.set('name', 'Alice')
    expect(() => s.toggle('name')).toThrow('non-boolean')
  })

  it('notifies subscribers', () => {
    const s = new PageState()
    const fn = vi.fn()
    s.subscribe('filters', fn)
    s.set('filters.month', 'July')
    expect(fn).toHaveBeenCalledWith('filters.month', 'July')
  })

  it('notifies on exact match', () => {
    const s = new PageState()
    const fn = vi.fn()
    s.subscribe('name', fn)
    s.set('name', 'Bob')
    expect(fn).toHaveBeenCalledWith('name', 'Bob')
  })

  it('does not notify unrelated subscribers', () => {
    const s = new PageState()
    const fn = vi.fn()
    s.subscribe('other', fn)
    s.set('name', 'Bob')
    expect(fn).not.toHaveBeenCalled()
  })

  it('unsubscribes correctly', () => {
    const s = new PageState()
    const fn = vi.fn()
    const unsub = s.subscribe('name', fn)
    unsub()
    s.set('name', 'Bob')
    expect(fn).not.toHaveBeenCalled()
  })

  it('wildcard subscriber gets all changes', () => {
    const s = new PageState()
    const fn = vi.fn()
    s.subscribe('*', fn)
    s.set('a', 1)
    s.set('b.c', 2)
    expect(fn).toHaveBeenCalledTimes(2)
  })

  it('snapshot returns deep copy', () => {
    const s = new PageState()
    s.set('data', { x: 1 })
    const snap = s.snapshot()
    snap.data = { x: 999 }
    expect(s.get('data')).toEqual({ x: 1 })
  })

  it('load replaces all state', () => {
    const s = new PageState()
    s.set('old', 'value')
    s.load({ new: 'data' })
    expect(s.get('old')).toBeUndefined()
    expect(s.get('new')).toBe('data')
  })
})
