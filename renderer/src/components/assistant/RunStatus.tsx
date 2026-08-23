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
      data-dashforge-component={instance.id}
      data-dashforge-type="assistant.run-status"
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        fontSize: '0.8rem',
        color: 'var(--dashforge-color-text-secondary, #94a3b8)',
        ...instance.style,
      }}
    >
      {children ?? (
        <>
          <span data-dashforge-slot="status-indicator" />
          <span data-dashforge-slot="status-label">
            <span data-dashforge-placeholder="run-status">Idle</span>
          </span>
          {props.showElapsed && <span data-dashforge-slot="elapsed" />}
          {props.showTokenCount && <span data-dashforge-slot="token-count" />}
        </>
      )}
    </div>
  )
}

AssistantRunStatus.displayName = 'assistant.run-status'
