import React from 'react'
import type { ComponentProps } from '../../registry'

export type RunState = 'idle' | 'running' | 'complete' | 'error' | 'cancelled'

export interface RunStatusProps {
  showElapsed?: boolean
  showTokenCount?: boolean
  compact?: boolean
}

export function AssistantRunStatus({ instance, children }: ComponentProps): React.ReactElement {
  const props = (instance.properties ?? {}) as unknown as RunStatusProps

  return (
    <div
      data-uiforge-component={instance.id}
      data-uiforge-type="assistant.run-status"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        fontSize: '0.8rem',
        color: 'var(--uiforge-color-text-secondary, #94a3b8)',
        ...instance.style,
      }}
    >
      {children ?? (
        <>
          <span data-uiforge-slot="status-indicator" />
          <span data-uiforge-slot="status-label">
            <span data-uiforge-placeholder="run-status">Idle</span>
          </span>
          {props.showElapsed && <span data-uiforge-slot="elapsed" />}
          {props.showTokenCount && <span data-uiforge-slot="token-count" />}
        </>
      )}
    </div>
  )
}

AssistantRunStatus.displayName = 'assistant.run-status'
