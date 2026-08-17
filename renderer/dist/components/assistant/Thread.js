import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
export function AssistantThread({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsxs("div", { "data-uiforge-component": instance.id, "data-uiforge-type": "assistant.thread", style: {
            display: 'flex',
            flexDirection: 'column',
            flex: 1,
            overflow: 'auto',
            ...instance.style,
        }, children: [children ?? (_jsx("div", { "data-uiforge-slot": "messages", style: { flex: 1 }, children: _jsx("span", { "data-uiforge-placeholder": "assistant.thread", children: "Thread messages render here via @assistant-ui/react Thread primitive" }) })), props.showToolCalls !== false && _jsx("span", { "data-uiforge-config": "showToolCalls", hidden: true }), props.streamingIndicator !== false && (_jsx("span", { "data-uiforge-config": "streamingIndicator", hidden: true }))] }));
}
AssistantThread.displayName = 'assistant.thread';
//# sourceMappingURL=Thread.js.map