import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
export function AssistantThreadList({ instance, children }) {
    const props = (instance.properties ?? {});
    return (_jsxs("nav", { "data-dashforge-component": instance.id, "data-dashforge-type": "assistant.thread-list", style: {
            display: 'flex',
            flexDirection: 'column',
            height: '100%',
            overflow: 'auto',
            ...instance.style,
        }, children: [props.showSearch !== false && (_jsx("div", { "data-dashforge-slot": "search", style: { padding: '8px' }, children: _jsx("span", { "data-dashforge-placeholder": "search", children: "Search threads" }) })), props.showNewButton !== false && (_jsx("div", { "data-dashforge-slot": "new-thread", style: { padding: '8px' }, children: _jsx("span", { "data-dashforge-placeholder": "new-thread", children: "New thread button" }) })), _jsx("div", { "data-dashforge-slot": "threads", style: { flex: 1, overflow: 'auto' }, children: children ?? (_jsx("span", { "data-dashforge-placeholder": "assistant.thread-list", children: "Thread list renders here from conversation data" })) })] }));
}
AssistantThreadList.displayName = 'assistant.thread-list';
//# sourceMappingURL=ThreadList.js.map