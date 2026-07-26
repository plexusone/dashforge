import React from 'react'
import type { ComponentProps } from '../../registry'

export interface ThreadListProps {
  showSearch?: boolean
  showNewButton?: boolean
  showFolders?: boolean
  showDelete?: boolean
}

export function AssistantThreadList({ instance, children }: ComponentProps): React.ReactElement {
  const props = (instance.properties ?? {}) as unknown as ThreadListProps

  return (
    <nav
      data-uiforge-component={instance.id}
      data-uiforge-type="assistant.thread-list"
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        overflow: 'auto',
        ...instance.style,
      }}
    >
      {props.showSearch !== false && (
        <div data-uiforge-slot="search" style={{ padding: '8px' }}>
          <span data-uiforge-placeholder="search">Search threads</span>
        </div>
      )}
      {props.showNewButton !== false && (
        <div data-uiforge-slot="new-thread" style={{ padding: '8px' }}>
          <span data-uiforge-placeholder="new-thread">New thread button</span>
        </div>
      )}
      <div data-uiforge-slot="threads" style={{ flex: 1, overflow: 'auto' }}>
        {children ?? (
          <span data-uiforge-placeholder="assistant.thread-list">
            Thread list renders here from conversation data
          </span>
        )}
      </div>
    </nav>
  )
}

AssistantThreadList.displayName = 'assistant.thread-list'
