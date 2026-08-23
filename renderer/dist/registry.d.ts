import type { ComponentType } from 'react';
import type { ComponentInstance } from './types';
export interface ComponentProps {
    instance: ComponentInstance;
    children?: React.ReactNode;
}
export type DashForgeComponent = ComponentType<ComponentProps>;
export declare function registerComponent(type: string, component: DashForgeComponent): void;
export declare function getComponent(type: string): DashForgeComponent | undefined;
export declare function hasComponent(type: string): boolean;
export declare function listComponents(): string[];
export declare function clearRegistry(): void;
//# sourceMappingURL=registry.d.ts.map