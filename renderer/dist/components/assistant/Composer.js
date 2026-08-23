import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
export function AssistantComposer({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsxs("div", { "data-dashforge-component": instance.id, "data-dashforge-type": "assistant.composer", style: {
            borderTop: '1px solid var(--dashforge-color-border, #e2e8f0)',
            padding: '16px',
            ...instance.style,
        }, children: [children ?? (_jsx("div", { "data-dashforge-slot": "input", children: _jsx("span", { "data-dashforge-placeholder": "assistant.composer", children: "Composer input renders here via @assistant-ui/react Composer primitive" }) })), props.placeholder && (_jsx("span", { "data-dashforge-config": "placeholder", "data-value": props.placeholder, hidden: true }))] }));
}
AssistantComposer.displayName = 'assistant.composer';
//# sourceMappingURL=Composer.js.map