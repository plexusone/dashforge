import { jsx as _jsx, Fragment as _Fragment, jsxs as _jsxs } from "react/jsx-runtime";
export function AssistantRunStatus({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsx("div", { "data-uiforge-component": instance.id, "data-uiforge-type": "assistant.run-status", style: {
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontSize: '0.8rem',
            color: 'var(--uiforge-color-text-secondary, #94a3b8)',
            ...instance.style,
        }, children: children ?? (_jsxs(_Fragment, { children: [_jsx("span", { "data-uiforge-slot": "status-indicator" }), _jsx("span", { "data-uiforge-slot": "status-label", children: _jsx("span", { "data-uiforge-placeholder": "run-status", children: "Idle" }) }), props.showElapsed && _jsx("span", { "data-uiforge-slot": "elapsed" }), props.showTokenCount && _jsx("span", { "data-uiforge-slot": "token-count" })] })) }));
}
AssistantRunStatus.displayName = 'assistant.run-status';
//# sourceMappingURL=RunStatus.js.map