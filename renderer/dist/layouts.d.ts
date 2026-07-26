import React from 'react';
import type { LayoutSpec, ComponentInstance } from './types';
interface LayoutProps {
    layout: LayoutSpec;
    components: ComponentInstance[];
    renderComponent: (instance: ComponentInstance) => React.ReactNode;
}
export declare function Layout({ layout, components, renderComponent }: LayoutProps): React.ReactElement;
export {};
//# sourceMappingURL=layouts.d.ts.map