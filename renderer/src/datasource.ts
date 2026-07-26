import type { Binding } from './types'
import { evaluateExpression, containsExpression } from './expression'

export interface DataSourceConnector {
  id: string
  execute(operation: string, params: Record<string, unknown>): Promise<unknown>
}

export class DataSourceRegistry {
  private connectors = new Map<string, DataSourceConnector>()

  register(connector: DataSourceConnector): void {
    this.connectors.set(connector.id, connector)
  }

  get(id: string): DataSourceConnector | undefined {
    return this.connectors.get(id)
  }

  has(id: string): boolean {
    return this.connectors.has(id)
  }

  async resolveBindings(
    bindings: Record<string, Binding>,
    ctx: Record<string, unknown>,
  ): Promise<Record<string, unknown>> {
    const results: Record<string, unknown> = {}

    const entries = Object.entries(bindings)
    const promises = entries.map(async ([key, binding]) => {
      const connector = this.connectors.get(binding.source)
      if (!connector) {
        results[key] = binding.default ?? null
        return
      }

      const resolvedParams: Record<string, unknown> = {}
      if (binding.parameters) {
        for (const [pKey, pValue] of Object.entries(binding.parameters)) {
          if (typeof pValue === 'string' && containsExpression(pValue)) {
            resolvedParams[pKey] = evaluateExpression(pValue, ctx)
          } else {
            resolvedParams[pKey] = pValue
          }
        }
      }

      let result = await connector.execute(binding.operation, resolvedParams)

      if (binding.transform && containsExpression(binding.transform)) {
        const transformCtx = { ...ctx, result }
        result = evaluateExpression(binding.transform, transformCtx)
      }

      results[key] = result
    })

    await Promise.all(promises)
    return results
  }
}
