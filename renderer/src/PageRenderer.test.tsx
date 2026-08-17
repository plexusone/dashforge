import React from 'react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PageRenderer } from './PageRenderer'
import { registerComponent, clearRegistry } from './registry'
import type { PageSpec, ComponentProps } from './index'

function StubCard({ instance }: ComponentProps) {
  return <div data-testid={`card-${instance.id}`}>{instance.properties?.title as string}</div>
}

function StubMetric({ instance }: ComponentProps) {
  return <div data-testid={`metric-${instance.id}`}>{instance.properties?.title as string}</div>
}

const basePage: PageSpec = {
  apiVersion: 'ui.plexusone.dev/v1',
  kind: 'Page',
  metadata: { id: 'test-page', name: 'test-page', title: 'Test Page' },
  profile: 'dashboard',
  layout: { type: 'responsive-grid', config: { columns: 12, gap: '8px' } },
  components: [
    {
      id: 'card-1',
      type: 'core.card',
      position: { row: 0, col: 0, colSpan: 6, rowSpan: 1 },
      properties: { title: 'Revenue' },
    },
    {
      id: 'metric-1',
      type: 'analytics.metric',
      position: { row: 0, col: 6, colSpan: 6, rowSpan: 1 },
      properties: { title: 'Orders' },
    },
  ],
}

describe('PageRenderer', () => {
  beforeEach(() => {
    clearRegistry()
    registerComponent('core.card', StubCard)
    registerComponent('analytics.metric', StubMetric)
  })

  it('renders a grid layout with components', () => {
    const { container } = render(<PageRenderer page={basePage} />)
    const grid = container.querySelector('[data-uiforge-layout="responsive-grid"]')
    expect(grid).not.toBeNull()
    expect(grid?.children.length).toBe(2)
  })

  it('sets page-level data attributes', () => {
    const { container } = render(<PageRenderer page={basePage} />)
    const root = container.querySelector('[data-uiforge-page="test-page"]')
    expect(root).not.toBeNull()
    expect(root?.getAttribute('data-uiforge-profile')).toBe('dashboard')
  })

  it('renders registered components', () => {
    render(<PageRenderer page={basePage} />)
    expect(screen.getByTestId('card-card-1')).toBeDefined()
    expect(screen.getByTestId('metric-metric-1')).toBeDefined()
  })

  it('shows placeholder for unregistered components', () => {
    clearRegistry()
    const { container } = render(<PageRenderer page={basePage} />)
    const missing = container.querySelectorAll('[data-uiforge-missing]')
    expect(missing.length).toBe(2)
  })

  it('hides components with visibility condition false', () => {
    const page: PageSpec = {
      ...basePage,
      components: [
        {
          id: 'hidden',
          type: 'core.card',
          visibility: { condition: 'false' },
        },
      ],
    }
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-uiforge-cell="hidden"]')).toBeNull()
  })

  it('applies theme tokens as CSS custom properties', () => {
    const page: PageSpec = {
      ...basePage,
      theme: { id: 'default', variant: 'dark', tokens: { 'color-primary': '#3b82f6' } },
    }
    const { container } = render(<PageRenderer page={page} />)
    const root = container.querySelector('[data-uiforge-page]') as HTMLElement
    expect(root.style.getPropertyValue('--uiforge-color-primary')).toBe('#3b82f6')
  })

  it('calls onError when a component throws', () => {
    function Broken(): React.ReactElement {
      throw new Error('boom')
    }
    clearRegistry()
    registerComponent('core.card', Broken)

    const onError = vi.fn()
    const page: PageSpec = {
      ...basePage,
      components: [{ id: 'bad', type: 'core.card' }],
    }

    const { container } = render(<PageRenderer page={page} onError={onError} />)
    expect(onError).toHaveBeenCalledWith('bad', expect.any(Error))
    expect(container.querySelector('[data-uiforge-error="bad"]')).not.toBeNull()
  })

  it('renders grid cell positioning', () => {
    const { container } = render(<PageRenderer page={basePage} />)
    const cell = container.querySelector('[data-uiforge-cell="card-1"]') as HTMLElement
    expect(cell.style.gridColumn).toBe('1 / span 6')
    expect(cell.style.gridRow).toBe('1 / span 1')
  })
})

describe('Stack layout', () => {
  beforeEach(() => {
    clearRegistry()
    registerComponent('core.card', StubCard)
  })

  it('renders a flex column by default', () => {
    const page: PageSpec = {
      ...basePage,
      layout: { type: 'stack' },
      components: [{ id: 'a', type: 'core.card', properties: { title: 'A' } }],
    }
    const { container } = render(<PageRenderer page={page} />)
    const stack = container.querySelector('[data-uiforge-layout="stack"]') as HTMLElement
    expect(stack.style.flexDirection).toBe('column')
  })
})

describe('Application shell layout', () => {
  beforeEach(() => {
    clearRegistry()
    registerComponent('core.card', StubCard)
  })

  it('renders main region with components', () => {
    const page: PageSpec = {
      ...basePage,
      layout: {
        type: 'application-shell',
        regions: [{ name: 'header' }, { name: 'sidebar' }, { name: 'main' }],
      },
      components: [{ id: 'content', type: 'core.card', properties: { title: 'Content' } }],
    }
    const { container } = render(<PageRenderer page={page} />)
    const main = container.querySelector('[data-uiforge-region="main"]')
    expect(main).not.toBeNull()
    expect(main?.querySelector('[data-testid="card-content"]')).not.toBeNull()
    expect(container.querySelector('[data-uiforge-region="header"]')).not.toBeNull()
    expect(container.querySelector('[data-uiforge-region="sidebar"]')).not.toBeNull()
  })
})
