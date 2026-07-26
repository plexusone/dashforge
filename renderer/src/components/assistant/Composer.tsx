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
      data-uiforge-component={instance.id}
      data-uiforge-type="assistant.composer"
      style={{
        borderTop: '1px solid var(--uiforge-color-border, #e2e8f0)',
        padding: '16px',
        ...instance.style,
      }}
    >
      {children ?? (
        <div data-uiforge-slot="input">
          <span data-uiforge-placeholder="assistant.composer">
            Composer input renders here via @assistant-ui/react Composer primitive
          </span>
        </div>
      )}
      {props.placeholder && (
        <span data-uiforge-config="placeholder" data-value={props.placeholder} hidden />
      )}
    </div>
  )
}

AssistantComposer.displayName = 'assistant.composer'
