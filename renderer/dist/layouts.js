import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import React from 'react';
export function Layout({ layout, components, renderComponent }) {
    switch (layout.type) {
        case 'responsive-grid':
            return (_jsx(GridLayout, { layout: layout, components: components, renderComponent: renderComponent }));
        case 'stack':
            return (_jsx(StackLayout, { layout: layout, components: components, renderComponent: renderComponent }));
        case 'split-pane':
            return _jsx(SplitPaneLayout, { layout: layout, renderComponent: renderComponent });
        case 'tabs':
            return _jsx(TabsLayout, { layout: layout, renderComponent: renderComponent });
        case 'application-shell':
            return (_jsx(AppShellLayout, { layout: layout, components: components, renderComponent: renderComponent }));
        default:
            return (_jsx("div", { "data-dashforge-error": `unknown layout: ${layout.type}`, children: components.map(renderComponent) }));
    }
}
function GridLayout({ layout, components, renderComponent }) {
    const columns = layout.config?.columns ?? 12;
    const gap = layout.config?.gap ?? '8px';
    const style = {
        display: 'grid',
        gridTemplateColumns: `repeat(${columns}, 1fr)`,
        gap,
    };
    return (_jsx("div", { style: style, "data-dashforge-layout": "responsive-grid", children: components.map((comp) => {
            const node = renderComponent(comp);
            if (node === null)
                return null;
            const pos = comp.position;
            const cellStyle = pos
                ? {
                    gridColumn: `${(pos.col ?? 0) + 1} / span ${pos.colSpan ?? 1}`,
                    gridRow: `${(pos.row ?? 0) + 1} / span ${pos.rowSpan ?? 1}`,
                }
                : {};
            return (_jsx("div", { style: cellStyle, "data-dashforge-cell": comp.id, children: node }, comp.id));
        }) }));
}
function StackLayout({ layout, components, renderComponent }) {
    const direction = layout.config?.direction ?? 'vertical';
    const gap = layout.config?.gap ?? '8px';
    const style = {
        display: 'flex',
        flexDirection: direction === 'horizontal' ? 'row' : 'column',
        gap,
    };
    return (_jsx("div", { style: style, "data-dashforge-layout": "stack", children: components.map((comp) => {
            const node = renderComponent(comp);
            if (node === null)
                return null;
            return (_jsx("div", { "data-dashforge-cell": comp.id, children: node }, comp.id));
        }) }));
}
function SplitPaneLayout({ layout, renderComponent }) {
    const direction = layout.config?.direction ?? 'horizontal';
    const sizes = layout.config?.sizes ?? [];
    const gap = layout.config?.gap ?? '0px';
    const regions = layout.regions ?? [];
    const style = {
        display: 'flex',
        flexDirection: direction === 'vertical' ? 'column' : 'row',
        gap,
        height: '100%',
    };
    return (_jsx("div", { style: style, "data-dashforge-layout": "split-pane", children: regions.map((region, i) => {
            const size = sizes[i];
            const paneStyle = size ? { flex: `0 0 ${size}` } : { flex: 1 };
            return (_jsx("div", { style: paneStyle, "data-dashforge-region": region.name, children: region.layout && (_jsx(Layout, { layout: region.layout, components: [], renderComponent: renderComponent })) }, region.name));
        }) }));
}
function TabsLayout({ layout }) {
    const regions = layout.regions ?? [];
    const [active, setActive] = React.useState(regions[0]?.name ?? '');
    return (_jsxs("div", { "data-dashforge-layout": "tabs", children: [_jsx("div", { role: "tablist", style: { display: 'flex', gap: '4px', borderBottom: '1px solid #e2e8f0' }, children: regions.map((region) => (_jsx("button", { role: "tab", "aria-selected": region.name === active, onClick: () => setActive(region.name), style: {
                        padding: '8px 16px',
                        border: 'none',
                        background: region.name === active ? '#e2e8f0' : 'transparent',
                        cursor: 'pointer',
                        fontWeight: region.name === active ? 600 : 400,
                    }, children: region.name }, region.name))) }), regions.map((region) => (_jsx("div", { role: "tabpanel", hidden: region.name !== active, "data-dashforge-region": region.name }, region.name)))] }));
}
function AppShellLayout({ layout, components, renderComponent }) {
    const regions = layout.regions ?? [];
    const regionMap = new Map(regions.map((r) => [r.name, r]));
    const style = {
        display: 'grid',
        gridTemplateRows: 'auto 1fr auto',
        gridTemplateColumns: 'auto 1fr auto',
        gridTemplateAreas: `
      "header header header"
      "sidebar main aside"
      "footer footer footer"
    `,
        minHeight: '100vh',
    };
    const slotStyle = (area) => ({ gridArea: area });
    return (_jsxs("div", { style: style, "data-dashforge-layout": "application-shell", children: [regionMap.has('header') && _jsx("div", { style: slotStyle('header'), "data-dashforge-region": "header" }), regionMap.has('sidebar') && (_jsx("div", { style: slotStyle('sidebar'), "data-dashforge-region": "sidebar" })), _jsx("div", { style: slotStyle('main'), "data-dashforge-region": "main", children: components.map(renderComponent) }), regionMap.has('aside') && _jsx("div", { style: slotStyle('aside'), "data-dashforge-region": "aside" }), regionMap.has('footer') && _jsx("div", { style: slotStyle('footer'), "data-dashforge-region": "footer" })] }));
}
//# sourceMappingURL=layouts.js.map