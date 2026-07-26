import { describe, it, expect } from 'vitest'
import { evaluateExpression, containsExpression, extractPaths } from './expression'

describe('evaluateExpression', () => {
  it('returns plain strings unchanged', () => {
    expect(evaluateExpression('hello world', {})).toBe('hello world')
  })

  it('resolves simple path', () => {
    const ctx = { context: { id: 'abc-123' } }
    expect(evaluateExpression('${context.id}', ctx)).toBe('abc-123')
  })

  it('resolves nested path', () => {
    const ctx = { state: { filters: { month: 'July' } } }
    expect(evaluateExpression('${state.filters.month}', ctx)).toBe('July')
  })

  it('preserves numeric type for full expression', () => {
    const ctx = { count: 42 }
    expect(evaluateExpression('${count}', ctx)).toBe(42)
  })

  it('preserves boolean type for full expression', () => {
    const ctx = { active: true }
    expect(evaluateExpression('${active}', ctx)).toBe(true)
  })

  it('preserves object type for full expression', () => {
    const obj = { a: 1, b: 2 }
    const ctx = { data: obj }
    expect(evaluateExpression('${data}', ctx)).toBe(obj)
  })

  it('interpolates mixed literal and expressions as string', () => {
    const ctx = { user: { name: 'Alice' } }
    expect(evaluateExpression('Hello ${user.name}!', ctx)).toBe('Hello Alice!')
  })

  it('interpolates multiple expressions', () => {
    const ctx = { first: 'Jane', last: 'Doe' }
    expect(evaluateExpression('${first} ${last}', ctx)).toBe('Jane Doe')
  })

  it('resolves array index', () => {
    const ctx = { items: [{ name: 'first' }, { name: 'second' }] }
    expect(evaluateExpression('${items.0.name}', ctx)).toBe('first')
    expect(evaluateExpression('${items.1.name}', ctx)).toBe('second')
  })

  it('throws on missing path', () => {
    expect(() => evaluateExpression('${missing.key}', {})).toThrow('not found')
  })

  it('throws on traversal into non-object', () => {
    const ctx = { name: 'hello' }
    expect(() => evaluateExpression('${name.length}', ctx)).toThrow('cannot traverse')
  })

  it('handles whitespace in expressions', () => {
    const ctx = { x: 'val' }
    expect(evaluateExpression('${ x }', ctx)).toBe('val')
  })
})

describe('containsExpression', () => {
  it('returns true for expressions', () => {
    expect(containsExpression('${foo}')).toBe(true)
    expect(containsExpression('hello ${bar} world')).toBe(true)
  })

  it('returns false for plain strings', () => {
    expect(containsExpression('hello world')).toBe(false)
    expect(containsExpression('$notexpr')).toBe(false)
    expect(containsExpression('')).toBe(false)
  })
})

describe('extractPaths', () => {
  it('extracts single path', () => {
    expect(extractPaths('${context.id}')).toEqual(['context.id'])
  })

  it('extracts multiple paths', () => {
    expect(extractPaths('${a.b} and ${c.d}')).toEqual(['a.b', 'c.d'])
  })

  it('returns empty for plain strings', () => {
    expect(extractPaths('no expressions here')).toEqual([])
  })

  it('trims whitespace from paths', () => {
    expect(extractPaths('${ foo.bar }')).toEqual(['foo.bar'])
  })
})
