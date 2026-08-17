import React from 'react';
import type { PageSpec } from './types';
import { PageState } from './state';
import { InteractionEngine } from './interaction';
import { DataSourceRegistry } from './datasource';
import type { DataSourceConnector } from './datasource';
export interface UIForgeContextValue {
    state: PageState;
    engine: InteractionEngine;
    dataSources: DataSourceRegistry;
    onInteraction?: (componentId: string, event: string, data?: Record<string, unknown>) => void;
}
export declare const UIForgeContext: React.Context<UIForgeContextValue | null>;
export declare function useUIForge(): UIForgeContextValue | null;
export interface PageRendererProps {
    page: PageSpec;
    className?: string;
    style?: React.CSSProperties;
    onError?: (componentId: string, error: Error) => void;
    initialState?: Record<string, unknown>;
    dataSources?: DataSourceConnector[];
    onInteraction?: (componentId: string, event: string, data?: Record<string, unknown>) => void;
}
export declare function PageRenderer({ page, className, style, onError, initialState, dataSources: dataSourceConnectors, onInteraction, }: PageRendererProps): React.ReactElement;
//# sourceMappingURL=PageRenderer.d.ts.map