import React from 'react'
import type { LayoutSpec, ComponentInstance } from './types'

interface LayoutProps {
  layout: LayoutSpec
  components: ComponentInstance[]
  renderComponent: (instance: ComponentInstance) => React.ReactNode
}

export function Layout({ layout, components, renderComponent }: LayoutProps): React.ReactElement {
  switch (layout.type) {
    case 'responsive-grid':
      return (
        <GridLayout layout={layout} components={components} renderComponent={renderComponent} />
      )
    case 'stack':
      return (
        <StackLayout layout={layout} components={components} renderComponent={renderComponent} />
      )
    case 'split-pane':
      return <SplitPaneLayout layout={layout} renderComponent={renderComponent} />
    case 'tabs':
      return <TabsLayout layout={layout} renderComponent={renderComponent} />
    case 'application-shell':
      return (
        <AppShellLayout layout={layout} components={components} renderComponent={renderComponent} />
      )
    default:
      return (
        <div data-uiforge-error={`unknown layout: ${layout.type}`}>
          {components.map(renderComponent)}
        </div>
      )
  }
}

function GridLayout({ layout, components, renderComponent }: LayoutProps): React.ReactElement {
  const columns = layout.config?.columns ?? 12
  const gap = layout.config?.gap ?? '8px'

  const style: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${columns}, 1fr)`,
    gap,
  }

  return (
    <div style={style} data-uiforge-layout="responsive-grid">
      {components.map((comp) => {
        const node = renderComponent(comp)
        if (node === null) return null
        const pos = comp.position
        const cellStyle: React.CSSProperties = pos
          ? {
              gridColumn: `${(pos.col ?? 0) + 1} / span ${pos.colSpan ?? 1}`,
              gridRow: `${(pos.row ?? 0) + 1} / span ${pos.rowSpan ?? 1}`,
            }
          : {}
        return (
          <div key={comp.id} style={cellStyle} data-uiforge-cell={comp.id}>
            {node}
          </div>
        )
      })}
    </div>
  )
}

function StackLayout({ layout, components, renderComponent }: LayoutProps): React.ReactElement {
  const direction = layout.config?.direction ?? 'vertical'
  const gap = layout.config?.gap ?? '8px'

  const style: React.CSSProperties = {
    display: 'flex',
    flexDirection: direction === 'horizontal' ? 'row' : 'column',
    gap,
  }

  return (
    <div style={style} data-uiforge-layout="stack">
      {components.map((comp) => {
        const node = renderComponent(comp)
        if (node === null) return null
        return (
          <div key={comp.id} data-uiforge-cell={comp.id}>
            {node}
          </div>
        )
      })}
    </div>
  )
}

interface RegionLayoutProps {
  layout: LayoutSpec
  renderComponent: (instance: ComponentInstance) => React.ReactNode
}

function SplitPaneLayout({ layout, renderComponent }: RegionLayoutProps): React.ReactElement {
  const direction = layout.config?.direction ?? 'horizontal'
  const sizes = layout.config?.sizes ?? []
  const gap = layout.config?.gap ?? '0px'
  const regions = layout.regions ?? []

  const style: React.CSSProperties = {
    display: 'flex',
    flexDirection: direction === 'vertical' ? 'column' : 'row',
    gap,
    height: '100%',
  }

  return (
    <div style={style} data-uiforge-layout="split-pane">
      {regions.map((region, i) => {
        const size = sizes[i]
        const paneStyle: React.CSSProperties = size ? { flex: `0 0 ${size}` } : { flex: 1 }
        return (
          <div key={region.name} style={paneStyle} data-uiforge-region={region.name}>
            {region.layout && (
              <Layout layout={region.layout} components={[]} renderComponent={renderComponent} />
            )}
          </div>
        )
      })}
    </div>
  )
}

function TabsLayout({ layout }: RegionLayoutProps): React.ReactElement {
  const regions = layout.regions ?? []
  const [active, setActive] = React.useState(regions[0]?.name ?? '')

  return (
    <div data-uiforge-layout="tabs">
      <div
        role="tablist"
        style={{ display: 'flex', gap: '4px', borderBottom: '1px solid #e2e8f0' }}
      >
        {regions.map((region) => (
          <button
            key={region.name}
            role="tab"
            aria-selected={region.name === active}
            onClick={() => setActive(region.name)}
            style={{
              padding: '8px 16px',
              border: 'none',
              background: region.name === active ? '#e2e8f0' : 'transparent',
              cursor: 'pointer',
              fontWeight: region.name === active ? 600 : 400,
            }}
          >
            {region.name}
          </button>
        ))}
      </div>
      {regions.map((region) => (
        <div
          key={region.name}
          role="tabpanel"
          hidden={region.name !== active}
          data-uiforge-region={region.name}
        />
      ))}
    </div>
  )
}

function AppShellLayout({ layout, components, renderComponent }: LayoutProps): React.ReactElement {
  const regions = layout.regions ?? []
  const regionMap = new Map(regions.map((r) => [r.name, r]))

  const style: React.CSSProperties = {
    display: 'grid',
    gridTemplateRows: 'auto 1fr auto',
    gridTemplateColumns: 'auto 1fr auto',
    gridTemplateAreas: `
      "header header header"
      "sidebar main aside"
      "footer footer footer"
    `,
    minHeight: '100vh',
  }

  const slotStyle = (area: string): React.CSSProperties => ({ gridArea: area })

  return (
    <div style={style} data-uiforge-layout="application-shell">
      {regionMap.has('header') && <div style={slotStyle('header')} data-uiforge-region="header" />}
      {regionMap.has('sidebar') && (
        <div style={slotStyle('sidebar')} data-uiforge-region="sidebar" />
      )}
      <div style={slotStyle('main')} data-uiforge-region="main">
        {components.map(renderComponent)}
      </div>
      {regionMap.has('aside') && <div style={slotStyle('aside')} data-uiforge-region="aside" />}
      {regionMap.has('footer') && <div style={slotStyle('footer')} data-uiforge-region="footer" />}
    </div>
  )
}
