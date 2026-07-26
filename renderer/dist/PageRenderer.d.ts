import React from 'react';
import type { PageSpec } from './types';
export interface PageRendererProps {
    page: PageSpec;
    className?: string;
    style?: React.CSSProperties;
    onError?: (componentId: string, error: Error) => void;
}
export declare function PageRenderer({ page, className, style, onError }: PageRendererProps): React.ReactElement;
//# sourceMappingURL=PageRenderer.d.ts.map