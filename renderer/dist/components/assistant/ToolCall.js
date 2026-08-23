import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
export function AssistantToolCall({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsx("div", { "data-dashforge-component": instance.id, "data-dashforge-type": "assistant.tool-call", style: {
            border: '1px solid var(--dashforge-color-border, #e2e8f0)',
            borderRadius: '8px',
            padding: '8px 12px',
            fontSize: '0.85rem',
            ...instance.style,
        }, children: children ?? (_jsxs(_Fragment, { children: [_jsxs("div", { "data-dashforge-slot": "tool-header", style: { display: 'flex', alignItems: 'center', gap: '8px' }, children: [_jsx("span", { "data-dashforge-placeholder": "tool-name", children: "Tool name" }), _jsx("span", { "data-dashforge-slot": "status-icon" })] }), props.showArgs !== false && (_jsx("div", { "data-dashforge-slot": "tool-args", style: { marginTop: '4px' }, children: _jsx("span", { "data-dashforge-placeholder": "tool-args", children: "Arguments" }) })), props.showResult !== false && (_jsx("div", { "data-dashforge-slot": "tool-result", style: { marginTop: '4px' }, children: _jsx("span", { "data-dashforge-placeholder": "tool-result", children: "Result" }) }))] })) }));
}
AssistantToolCall.displayName = 'assistant.tool-call';
//# sourceMappingURL=ToolCall.js.map