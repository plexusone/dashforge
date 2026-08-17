import { evaluateExpression, containsExpression } from './expression';
export class InteractionEngine {
    constructor(state) {
        this.handlers = new Map();
        this.state = state;
        this.registerBuiltins();
    }
    registerHandler(action, handler) {
        this.handlers.set(action, handler);
    }
    dispatch(interactions, componentId, eventName, eventData) {
        const data = eventData ?? {};
        for (const interaction of interactions) {
            if (interaction.when.component !== componentId || interaction.when.event !== eventName) {
                continue;
            }
            for (const action of interaction.then) {
                if (action.params?.condition) {
                    const condStr = String(action.params.condition);
                    if (containsExpression(condStr)) {
                        const ctx = { event: data, state: this.state.snapshot() };
                        const result = evaluateExpression(condStr, ctx);
                        if (!result)
                            continue;
                    }
                    else if (condStr === 'false') {
                        continue;
                    }
                }
                const handler = this.handlers.get(action.action);
                if (handler) {
                    handler(action, data);
                }
            }
        }
    }
    registerBuiltins() {
        this.registerHandler('state.set', (action, eventData) => {
            const path = action.params?.path;
            if (!path)
                return;
            let value = action.params?.value;
            if (typeof value === 'string' && containsExpression(value)) {
                const ctx = { event: eventData, state: this.state.snapshot() };
                value = evaluateExpression(value, ctx);
            }
            this.state.set(path, value);
        });
        this.registerHandler('state.toggle', (action) => {
            const path = action.params?.path;
            if (!path)
                return;
            this.state.toggle(path);
        });
        this.registerHandler('component.refresh', () => {
            // Placeholder — in a full runtime this would trigger a re-fetch for the target component
        });
        this.registerHandler('navigate', () => {
            // Placeholder — in a full runtime this would trigger navigation
        });
    }
}
//# sourceMappingURL=interaction.js.map