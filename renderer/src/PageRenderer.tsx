import React from 'react'
import type { PageSpec, ComponentInstance, ThemeRef } from './types'
import { getComponent } from './registry'
import { Layout } from './layouts'
import { evaluateExpression, containsExpression } from './expression'
import { PageState } from './state'
import { InteractionEngine } from './interaction'
import { DataSourceRegistry } from './datasource'
import type { DataSourceConnector } from './datasource'

export interface DashForgeContextValue {
  state: PageState
  engine: InteractionEngine
  dataSources: DataSourceRegistry
  onInteraction?: (componentId: string, event: string, data?: Record<string, unknown>) => void
}

export const DashForgeContext = React.createContext<DashForgeContextValue | null>(null)

export function useDashForge(): DashForgeContextValue | null {
  return React.useContext(DashForgeContext)
}

export interface PageRendererProps {
  page: PageSpec
  className?: string
  style?: React.CSSProperties
  onError?: (componentId: string, error: Error) => void
  initialState?: Record<string, unknown>
  dataSources?: DataSourceConnector[]
  onInteraction?: (componentId: string, event: string, data?: Record<string, unknown>) => void
}

export function PageRenderer({
  page,
  className,
  style,
  onError,
  initialState,
  dataSources: dataSourceConnectors,
  onInteraction,
}: PageRendererProps): React.ReactElement {
  const [ctx] = React.useState<DashForgeContextValue>(() => {
    const pageState = new PageState()
    if (page.context) {
      pageState.load({ context: page.context })
    }
    if (initialState) {
      const merged = { ...pageState.snapshot(), ...initialState }
      pageState.load(merged)
    }
    const engine = new InteractionEngine(pageState)
    const dsRegistry = new DataSourceRegistry()
    if (dataSourceConnectors) {
      for (const c of dataSourceConnectors) {
        dsRegistry.register(c)
      }
    }
    return { state: pageState, engine, dataSources: dsRegistry, onInteraction }
  })

  const themeStyle = buildThemeStyle(page.theme)
  const mergedStyle = { ...themeStyle, ...style }

  function renderComponent(instance: ComponentInstance): React.ReactNode {
    if (instance.visibility?.condition) {
      const cond = instance.visibility.condition
      if (cond === 'false') {
        return null
      }
      if (containsExpression(cond)) {
        const exprCtx = { state: ctx.state.snapshot(), context: page.context ?? {} }
        try {
          const result = evaluateExpression(cond, exprCtx)
          if (!result) return null
        } catch {
          return null
        }
      }
    }

    const Component = getComponent(instance.type)
    if (!Component) {
      return (
        <div
          key={instance.id}
          data-dashforge-missing={instance.type}
          style={{
            padding: '8px',
            border: '1px dashed #cbd5e1',
            borderRadius: '4px',
            color: '#94a3b8',
            fontSize: '0.8rem',
          }}
        >
          Unknown component: {instance.type}
        </div>
      )
    }

    const children = instance.children?.map(renderComponent)

    return (
      <ErrorBoundary key={instance.id} componentId={instance.id} onError={onError}>
        <Component instance={instance}>{children}</Component>
      </ErrorBoundary>
    )
  }

  return (
    <DashForgeContext.Provider value={ctx}>
      <div
        className={className}
        style={mergedStyle}
        data-dashforge-page={page.metadata.id}
        data-dashforge-profile={page.profile}
      >
        <Layout
          layout={page.layout}
          components={page.components}
          renderComponent={renderComponent}
        />
      </div>
    </DashForgeContext.Provider>
  )
}

function buildThemeStyle(theme?: ThemeRef): React.CSSProperties {
  if (!theme?.tokens) return {}
  const style: Record<string, string> = {}
  for (const [key, value] of Object.entries(theme.tokens)) {
    style[`--dashforge-${key}`] = value
  }
  return style
}

interface ErrorBoundaryProps {
  componentId: string
  onError?: (componentId: string, error: Error) => void
  children: React.ReactNode
}

interface ErrorBoundaryState {
  error: Error | null
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  componentDidCatch(error: Error): void {
    this.props.onError?.(this.props.componentId, error)
  }

  render(): React.ReactNode {
    if (this.state.error) {
      return (
        <div
          data-dashforge-error={this.props.componentId}
          style={{
            padding: '8px',
            border: '1px solid #ef4444',
            borderRadius: '4px',
            color: '#ef4444',
            fontSize: '0.8rem',
          }}
        >
          Error in {this.props.componentId}: {this.state.error.message}
        </div>
      )
    }
    return this.props.children
  }
}
