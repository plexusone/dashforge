import { evaluateExpression, containsExpression } from './expression';
export class DataSourceRegistry {
    constructor() {
        this.connectors = new Map();
    }
    register(connector) {
        this.connectors.set(connector.id, connector);
    }
    get(id) {
        return this.connectors.get(id);
    }
    has(id) {
        return this.connectors.has(id);
    }
    async resolveBindings(bindings, ctx) {
        const results = {};
        const entries = Object.entries(bindings);
        const promises = entries.map(async ([key, binding]) => {
            const connector = this.connectors.get(binding.source);
            if (!connector) {
                results[key] = binding.default ?? null;
                return;
            }
            const resolvedParams = {};
            if (binding.parameters) {
                for (const [pKey, pValue] of Object.entries(binding.parameters)) {
                    if (typeof pValue === 'string' && containsExpression(pValue)) {
                        resolvedParams[pKey] = evaluateExpression(pValue, ctx);
                    }
                    else {
                        resolvedParams[pKey] = pValue;
                    }
                }
            }
            let result = await connector.execute(binding.operation, resolvedParams);
            if (binding.transform && containsExpression(binding.transform)) {
                const transformCtx = { ...ctx, result };
                result = evaluateExpression(binding.transform, transformCtx);
            }
            results[key] = result;
        });
        await Promise.all(promises);
        return results;
    }
}
//# sourceMappingURL=datasource.js.map