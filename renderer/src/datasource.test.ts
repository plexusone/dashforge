import { describe, it, expect, vi } from 'vitest'
import { DataSourceRegistry } from './datasource'
import type { DataSourceConnector } from './datasource'
import type { Binding } from './types'

function makeConnector(id: string, data: Record<string, unknown>): DataSourceConnector {
  return {
    id,
    execute: vi.fn(async (operation: string, params: Record<string, unknown>) => {
      return data[operation] ?? params
    }),
  }
}

describe('DataSourceRegistry', () => {
  it('registers and retrieves connectors', () => {
    const reg = new DataSourceRegistry()
    const c = makeConnector('api', {})
    reg.register(c)
    expect(reg.get('api')).toBe(c)
    expect(reg.has('api')).toBe(true)
    expect(reg.has('other')).toBe(false)
  })

  it('resolves bindings with connector', async () => {
    const reg = new DataSourceRegistry()
    const c = makeConnector('customer-api', { getCustomer: { name: 'Alice' } })
    reg.register(c)

    const bindings: Record<string, Binding> = {
      customer: {
        source: 'customer-api',
        operation: 'getCustomer',
        parameters: { id: '123' },
      },
    }

    const result = await reg.resolveBindings(bindings, {})
    expect(result.customer).toEqual({ name: 'Alice' })
    expect(c.execute).toHaveBeenCalledWith('getCustomer', { id: '123' })
  })

  it('evaluates expression parameters', async () => {
    const reg = new DataSourceRegistry()
    const c = makeConnector('api', { getData: 'ok' })
    reg.register(c)

    const bindings: Record<string, Binding> = {
      data: {
        source: 'api',
        operation: 'getData',
        parameters: { id: '${context.entityId}' },
      },
    }

    const ctx = { context: { entityId: 'ent-42' } }
    await reg.resolveBindings(bindings, ctx)
    expect(c.execute).toHaveBeenCalledWith('getData', { id: 'ent-42' })
  })

  it('falls back to default when connector is missing', async () => {
    const reg = new DataSourceRegistry()

    const bindings: Record<string, Binding> = {
      missing: {
        source: 'nonexistent',
        operation: 'op',
        default: 'fallback',
      },
    }

    const result = await reg.resolveBindings(bindings, {})
    expect(result.missing).toBe('fallback')
  })

  it('falls back to null when connector is missing and no default', async () => {
    const reg = new DataSourceRegistry()

    const bindings: Record<string, Binding> = {
      missing: {
        source: 'nonexistent',
        operation: 'op',
      },
    }

    const result = await reg.resolveBindings(bindings, {})
    expect(result.missing).toBeNull()
  })

  it('resolves multiple bindings in parallel', async () => {
    const reg = new DataSourceRegistry()
    const c1 = makeConnector('api-a', { getA: 'valueA' })
    const c2 = makeConnector('api-b', { getB: 'valueB' })
    reg.register(c1)
    reg.register(c2)

    const bindings: Record<string, Binding> = {
      a: { source: 'api-a', operation: 'getA' },
      b: { source: 'api-b', operation: 'getB' },
    }

    const result = await reg.resolveBindings(bindings, {})
    expect(result.a).toBe('valueA')
    expect(result.b).toBe('valueB')
  })
})
