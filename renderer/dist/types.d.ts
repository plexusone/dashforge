export declare const API_VERSION = "ui.plexusone.dev/v1";
export declare const KIND_PAGE = "Page";
export interface PageSpec {
    apiVersion: string;
    kind: string;
    metadata: PageMetadata;
    profile?: string;
    context?: Record<string, string>;
    layout: LayoutSpec;
    components: ComponentInstance[];
    interactions?: Interaction[];
    navigation?: NavigationSpec;
    theme?: ThemeRef;
}
export interface PageMetadata {
    id: string;
    name: string;
    title: string;
    description?: string;
    version?: string;
    labels?: Record<string, string>;
}
export type LayoutType = 'responsive-grid' | 'stack' | 'split-pane' | 'tabs' | 'application-shell';
export interface LayoutSpec {
    type: LayoutType;
    config?: LayoutConfig;
    regions?: LayoutRegion[];
}
export interface LayoutConfig {
    columns?: number;
    rows?: number;
    gap?: string;
    direction?: 'horizontal' | 'vertical';
    breakpoints?: Record<string, BreakpointConfig>;
    sizes?: string[];
    resizable?: boolean;
}
export interface BreakpointConfig {
    columns: number;
    gap?: string;
}
export interface LayoutRegion {
    name: string;
    layout?: LayoutSpec;
}
export interface ComponentInstance {
    id: string;
    type: string;
    version?: string;
    position?: Position;
    properties?: Record<string, unknown>;
    data?: Record<string, Binding>;
    children?: ComponentInstance[];
    visibility?: VisibilityRule;
    slot?: string;
    style?: Record<string, string>;
    rawConfig?: unknown;
}
export interface Position {
    row?: number;
    col?: number;
    rowSpan?: number;
    colSpan?: number;
    order?: number;
    region?: string;
}
export interface Binding {
    source: string;
    operation: string;
    parameters?: Record<string, unknown>;
    transform?: string;
    default?: unknown;
}
export interface VisibilityRule {
    condition?: string;
    roles?: string[];
    capability?: string;
}
export interface Interaction {
    when: InteractionTrigger;
    then: InteractionAction[];
}
export interface InteractionTrigger {
    component: string;
    event: string;
}
export interface InteractionAction {
    target: string;
    action: string;
    params?: Record<string, unknown>;
}
export interface NavigationSpec {
    type: string;
    items: NavItem[];
}
export interface NavItem {
    id: string;
    label: string;
    icon?: string;
    target?: string;
    children?: NavItem[];
}
export interface ThemeRef {
    id: string;
    variant?: string;
    tokens?: Record<string, string>;
}
//# sourceMappingURL=types.d.ts.map