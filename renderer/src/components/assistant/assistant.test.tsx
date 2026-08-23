import { describe, it, expect, beforeEach } from 'vitest'
import { render } from '@testing-library/react'
import { PageRenderer } from '../../PageRenderer'
import { clearRegistry } from '../../registry'
import { registerAssistantComponents } from './register'
import type { PageSpec } from '../../types'

function agentPage(components: PageSpec['components']): PageSpec {
  return {
    apiVersion: 'ui.plexusone.dev/v1',
    kind: 'Page',
    metadata: { id: 'test-agent', name: 'test-agent', title: 'Agent Test' },
    layout: {
      type: 'stack',
      config: { direction: 'vertical' },
    },
    components,
  }
}

describe('Assistant components', () => {
  beforeEach(() => {
    clearRegistry()
    registerAssistantComponents()
  })

  it('registers all 5 assistant components', () => {
    const { components } = registerAssistantComponents()
    expect(components).toEqual([
      'assistant.thread',
      'assistant.composer',
      'assistant.thread-list',
      'assistant.tool-call',
      'assistant.run-status',
    ])
  })

  it('renders assistant.thread', () => {
    const page = agentPage([
      { id: 'thread', type: 'assistant.thread', properties: { showToolCalls: true } },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-type="assistant.thread"]')).not.toBeNull()
  })

  it('renders assistant.composer', () => {
    const page = agentPage([
      {
        id: 'composer',
        type: 'assistant.composer',
        properties: { placeholder: 'Ask something...' },
      },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-type="assistant.composer"]')).not.toBeNull()
  })

  it('renders assistant.thread-list with search slot', () => {
    const page = agentPage([
      { id: 'threads', type: 'assistant.thread-list', properties: { showSearch: true } },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-type="assistant.thread-list"]')).not.toBeNull()
    expect(container.querySelector('[data-dashforge-slot="search"]')).not.toBeNull()
  })

  it('renders assistant.tool-call', () => {
    const page = agentPage([
      {
        id: 'tool',
        type: 'assistant.tool-call',
        properties: { showArgs: true, collapsible: true },
      },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-type="assistant.tool-call"]')).not.toBeNull()
    expect(container.querySelector('[data-dashforge-slot="tool-header"]')).not.toBeNull()
  })

  it('renders assistant.run-status', () => {
    const page = agentPage([
      { id: 'status', type: 'assistant.run-status', properties: { showElapsed: true } },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-type="assistant.run-status"]')).not.toBeNull()
    expect(container.querySelector('[data-dashforge-slot="elapsed"]')).not.toBeNull()
  })

  it('hides search slot when showSearch is false', () => {
    const page = agentPage([
      { id: 'threads', type: 'assistant.thread-list', properties: { showSearch: false } },
    ])
    const { container } = render(<PageRenderer page={page} />)
    expect(container.querySelector('[data-dashforge-slot="search"]')).toBeNull()
  })
})
