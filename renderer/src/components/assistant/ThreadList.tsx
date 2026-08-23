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
      data-dashforge-component={instance.id}
      data-dashforge-type="assistant.thread-list"
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        overflow: 'auto',
        ...instance.style,
      }}
    >
      {props.showSearch !== false && (
        <div data-dashforge-slot="search" style={{ padding: '8px' }}>
          <span data-dashforge-placeholder="search">Search threads</span>
        </div>
      )}
      {props.showNewButton !== false && (
        <div data-dashforge-slot="new-thread" style={{ padding: '8px' }}>
          <span data-dashforge-placeholder="new-thread">New thread button</span>
        </div>
      )}
      <div data-dashforge-slot="threads" style={{ flex: 1, overflow: 'auto' }}>
        {children ?? (
          <span data-dashforge-placeholder="assistant.thread-list">
            Thread list renders here from conversation data
          </span>
        )}
      </div>
    </nav>
  )
}

AssistantThreadList.displayName = 'assistant.thread-list'
