import type { Interaction, InteractionAction } from './types';
import { PageState } from './state';
export type ActionHandler = (action: InteractionAction, eventData: Record<string, unknown>) => void;
export declare class InteractionEngine {
    private handlers;
    private state;
    constructor(state: PageState);
    registerHandler(action: string, handler: ActionHandler): void;
    dispatch(interactions: Interaction[], componentId: string, eventName: string, eventData?: Record<string, unknown>): void;
    private registerBuiltins;
}
//# sourceMappingURL=interaction.d.ts.map