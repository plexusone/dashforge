type Listener = (key: string, value: unknown) => void;
export declare class PageState {
    private values;
    private listeners;
    get(path: string): unknown;
    set(path: string, value: unknown): void;
    toggle(path: string): void;
    subscribe(pathPrefix: string, fn: Listener): () => void;
    snapshot(): Record<string, unknown>;
    load(initial: Record<string, unknown>): void;
    private notify;
}
export {};
//# sourceMappingURL=state.d.ts.map