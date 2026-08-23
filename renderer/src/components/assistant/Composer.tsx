import React from 'react'
import type { ComponentProps } from '../../registry'

export interface ComposerProps {
  placeholder?: string
  showAttachments?: boolean
  showModelSelector?: boolean
  maxLength?: number
}

export function AssistantComposer({ instance, children }: ComponentProps): React.ReactElement {
  const props = (instance.properties ?? {}) as unknown as ComposerProps

  return (
    <div
      data-dashforge-component={instance.id}
      data-dashforge-type="assistant.composer"
      style={{
        borderTop: '1px solid var(--dashforge-color-border, #e2e8f0)',
        padding: '16px',
        ...instance.style,
      }}
    >
      {children ?? (
        <div data-dashforge-slot="input">
          <span data-dashforge-placeholder="assistant.composer">
            Composer input renders here via @assistant-ui/react Composer primitive
          </span>
        </div>
      )}
      {props.placeholder && (
        <span data-dashforge-config="placeholder" data-value={props.placeholder} hidden />
      )}
    </div>
  )
}

AssistantComposer.displayName = 'assistant.composer'
