import type { Interaction, InteractionAction } from './types'
import { evaluateExpression, containsExpression } from './expression'
import { PageState } from './state'

export type ActionHandler = (action: InteractionAction, eventData: Record<string, unknown>) => void

export class InteractionEngine {
  private handlers = new Map<string, ActionHandler>()
  private state: PageState

  constructor(state: PageState) {
    this.state = state
    this.registerBuiltins()
  }

  registerHandler(action: string, handler: ActionHandler): void {
    this.handlers.set(action, handler)
  }

  dispatch(
    interactions: Interaction[],
    componentId: string,
    eventName: string,
    eventData?: Record<string, unknown>,
  ): void {
    const data = eventData ?? {}
    for (const interaction of interactions) {
      if (interaction.when.component !== componentId || interaction.when.event !== eventName) {
        continue
      }
      for (const action of interaction.then) {
        if (action.params?.condition) {
          const condStr = String(action.params.condition)
          if (containsExpression(condStr)) {
            const ctx = { event: data, state: this.state.snapshot() }
            const result = evaluateExpression(condStr, ctx)
            if (!result) continue
          } else if (condStr === 'false') {
            continue
          }
        }

        const handler = this.handlers.get(action.action)
        if (handler) {
          handler(action, data)
        }
      }
    }
  }

  private registerBuiltins(): void {
    this.registerHandler('state.set', (action, eventData) => {
      const path = action.params?.path as string | undefined
      if (!path) return
      let value = action.params?.value
      if (typeof value === 'string' && containsExpression(value)) {
        const ctx = { event: eventData, state: this.state.snapshot() }
        value = evaluateExpression(value, ctx)
      }
      this.state.set(path, value)
    })

    this.registerHandler('state.toggle', (action) => {
      const path = action.params?.path as string | undefined
      if (!path) return
      this.state.toggle(path)
    })

    this.registerHandler('component.refresh', () => {
      // Placeholder — in a full runtime this would trigger a re-fetch for the target component
    })

    this.registerHandler('navigate', () => {
      // Placeholder — in a full runtime this would trigger navigation
    })
  }
}
