import React from 'react'
import type { ComponentProps } from '../../registry'

export interface ThreadProps {
  showToolCalls?: boolean
  showTimestamps?: boolean
  markdownEnabled?: boolean
  streamingIndicator?: boolean
}

export function AssistantThread({ instance, children }: ComponentProps): React.ReactElement {
  const props = (instance.properties ?? {}) as unknown as ThreadProps

  return (
    <div
      data-uiforge-component={instance.id}
      data-uiforge-type="assistant.thread"
      style={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        overflow: 'auto',
        ...instance.style,
      }}
    >
      {children ?? (
        <div data-uiforge-slot="messages" style={{ flex: 1 }}>
          <span data-uiforge-placeholder="assistant.thread">
            Thread messages render here via @assistant-ui/react Thread primitive
          </span>
        </div>
      )}
      {props.showToolCalls !== false && <span data-uiforge-config="showToolCalls" hidden />}
      {props.streamingIndicator !== false && (
        <span data-uiforge-config="streamingIndicator" hidden />
      )}
    </div>
  )
}

AssistantThread.displayName = 'assistant.thread'
