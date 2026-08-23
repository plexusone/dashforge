import { jsx as _jsx, Fragment as _Fragment, jsxs as _jsxs } from "react/jsx-runtime";
export function AssistantRunStatus({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsx("div", { "data-dashforge-component": instance.id, "data-dashforge-type": "assistant.run-status", style: {
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            fontSize: '0.8rem',
            color: 'var(--dashforge-color-text-secondary, #94a3b8)',
            ...instance.style,
        }, children: children ?? (_jsxs(_Fragment, { children: [_jsx("span", { "data-dashforge-slot": "status-indicator" }), _jsx("span", { "data-dashforge-slot": "status-label", children: _jsx("span", { "data-dashforge-placeholder": "run-status", children: "Idle" }) }), props.showElapsed && _jsx("span", { "data-dashforge-slot": "elapsed" }), props.showTokenCount && _jsx("span", { "data-dashforge-slot": "token-count" })] })) }));
}
AssistantRunStatus.displayName = 'assistant.run-status';
//# sourceMappingURL=RunStatus.js.map