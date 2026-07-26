import { jsxs as _jsxs, jsx as _jsx } from "react/jsx-runtime";
import React from 'react';
import { getComponent } from './registry';
import { Layout } from './layouts';
export function PageRenderer({ page, className, style, onError }) {
    const themeStyle = buildThemeStyle(page.theme);
    const mergedStyle = { ...themeStyle, ...style };
    const renderComponent = React.useCallback((instance) => {
        if (instance.visibility?.condition === 'false') {
            return null;
        }
        const Component = getComponent(instance.type);
        if (!Component) {
            return (_jsxs("div", { "data-uiforge-missing": instance.type, style: { padding: '8px', border: '1px dashed #cbd5e1', borderRadius: '4px', color: '#94a3b8', fontSize: '0.8rem' }, children: ["Unknown component: ", instance.type] }, instance.id));
        }
        const children = instance.children?.map(renderComponent);
        return (_jsx(ErrorBoundary, { componentId: instance.id, onError: onError, children: _jsx(Component, { instance: instance, children: children }) }, instance.id));
    }, [onError]);
    return (_jsx("div", { className: className, style: mergedStyle, "data-uiforge-page": page.metadata.id, "data-uiforge-profile": page.profile, children: _jsx(Layout, { layout: page.layout, components: page.components, renderComponent: renderComponent }) }));
}
function buildThemeStyle(theme) {
    if (!theme?.tokens)
        return {};
    const style = {};
    for (const [key, value] of Object.entries(theme.tokens)) {
        style[`--uiforge-${key}`] = value;
    }
    return style;
}
class ErrorBoundary extends React.Component {
    constructor() {
        super(...arguments);
        this.state = { error: null };
    }
    static getDerivedStateFromError(error) {
        return { error };
    }
    componentDidCatch(error) {
        this.props.onError?.(this.props.componentId, error);
    }
    render() {
        if (this.state.error) {
            return (_jsxs("div", { "data-uiforge-error": this.props.componentId, style: { padding: '8px', border: '1px solid #ef4444', borderRadius: '4px', color: '#ef4444', fontSize: '0.8rem' }, children: ["Error in ", this.props.componentId, ": ", this.state.error.message] }));
        }
        return this.props.children;
    }
}
//# sourceMappingURL=PageRenderer.js.map