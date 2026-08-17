import React from 'react'
import type { ComponentProps } from '../../registry'

export interface ToolCallProps {
  showArgs?: boolean
  showResult?: boolean
  collapsible?: boolean
}

export function AssistantToolCall({ instance, children }: ComponentProps): React.ReactElement {
  const props = (instance.properties ?? {}) as unknown as ToolCallProps

  return (
    <div
      data-uiforge-component={instance.id}
      data-uiforge-type="assistant.tool-call"
      style={{
        border: '1px solid var(--uiforge-color-border, #e2e8f0)',
        borderRadius: '8px',
        padding: '8px 12px',
        fontSize: '0.85rem',
        ...instance.style,
      }}
    >
      {children ?? (
        <>
          <div
            data-uiforge-slot="tool-header"
            style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
          >
            <span data-uiforge-placeholder="tool-name">Tool name</span>
            <span data-uiforge-slot="status-icon" />
          </div>
          {props.showArgs !== false && (
            <div data-uiforge-slot="tool-args" style={{ marginTop: '4px' }}>
              <span data-uiforge-placeholder="tool-args">Arguments</span>
            </div>
          )}
          {props.showResult !== false && (
            <div data-uiforge-slot="tool-result" style={{ marginTop: '4px' }}>
              <span data-uiforge-placeholder="tool-result">Result</span>
            </div>
          )}
        </>
      )}
    </div>
  )
}

AssistantToolCall.displayName = 'assistant.tool-call'
