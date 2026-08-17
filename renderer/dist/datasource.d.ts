import type { Binding } from './types';
export interface DataSourceConnector {
    id: string;
    execute(operation: string, params: Record<string, unknown>): Promise<unknown>;
}
export declare class DataSourceRegistry {
    private connectors;
    register(connector: DataSourceConnector): void;
    get(id: string): DataSourceConnector | undefined;
    has(id: string): boolean;
    resolveBindings(bindings: Record<string, Binding>, ctx: Record<string, unknown>): Promise<Record<string, unknown>>;
}
//# sourceMappingURL=datasource.d.ts.map