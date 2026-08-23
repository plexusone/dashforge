import { jsxs as _jsxs, jsx as _jsx } from "react/jsx-runtime";
import React from 'react';
import { getComponent } from './registry';
import { Layout } from './layouts';
import { evaluateExpression, containsExpression } from './expression';
import { PageState } from './state';
import { InteractionEngine } from './interaction';
import { DataSourceRegistry } from './datasource';
export const DashForgeContext = React.createContext(null);
export function useDashForge() {
    return React.useContext(DashForgeContext);
}
export function PageRenderer({ page, className, style, onError, initialState, dataSources: dataSourceConnectors, onInteraction, }) {
    const [ctx] = React.useState(() => {
        const pageState = new PageState();
        if (page.context) {
            pageState.load({ context: page.context });
        }
        if (initialState) {
            const merged = { ...pageState.snapshot(), ...initialState };
            pageState.load(merged);
        }
        const engine = new InteractionEngine(pageState);
        const dsRegistry = new DataSourceRegistry();
        if (dataSourceConnectors) {
            for (const c of dataSourceConnectors) {
                dsRegistry.register(c);
            }
        }
        return { state: pageState, engine, dataSources: dsRegistry, onInteraction };
    });
    const themeStyle = buildThemeStyle(page.theme);
    const mergedStyle = { ...themeStyle, ...style };
    function renderComponent(instance) {
        if (instance.visibility?.condition) {
            const cond = instance.visibility.condition;
            if (cond === 'false') {
                return null;
            }
            if (containsExpression(cond)) {
                const exprCtx = { state: ctx.state.snapshot(), context: page.context ?? {} };
                try {
                    const result = evaluateExpression(cond, exprCtx);
                    if (!result)
                        return null;
                }
                catch {
                    return null;
                }
            }
        }
        const Component = getComponent(instance.type);
        if (!Component) {
            return (_jsxs("div", { "data-dashforge-missing": instance.type, style: {
                    padding: '8px',
                    border: '1px dashed #cbd5e1',
                    borderRadius: '4px',
                    color: '#94a3b8',
                    fontSize: '0.8rem',
                }, children: ["Unknown component: ", instance.type] }, instance.id));
        }
        const children = instance.children?.map(renderComponent);
        return (_jsx(ErrorBoundary, { componentId: instance.id, onError: onError, children: _jsx(Component, { instance: instance, children: children }) }, instance.id));
    }
    return (_jsx(DashForgeContext.Provider, { value: ctx, children: _jsx("div", { className: className, style: mergedStyle, "data-dashforge-page": page.metadata.id, "data-dashforge-profile": page.profile, children: _jsx(Layout, { layout: page.layout, components: page.components, renderComponent: renderComponent }) }) }));
}
function buildThemeStyle(theme) {
    if (!theme?.tokens)
        return {};
    const style = {};
    for (const [key, value] of Object.entries(theme.tokens)) {
        style[`--dashforge-${key}`] = value;
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
            return (_jsxs("div", { "data-dashforge-error": this.props.componentId, style: {
                    padding: '8px',
                    border: '1px solid #ef4444',
                    borderRadius: '4px',
                    color: '#ef4444',
                    fontSize: '0.8rem',
                }, children: ["Error in ", this.props.componentId, ": ", this.state.error.message] }));
        }
        return this.props.children;
    }
}
//# sourceMappingURL=PageRenderer.js.map