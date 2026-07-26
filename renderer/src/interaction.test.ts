import { describe, it, expect, vi } from 'vitest'
import { InteractionEngine } from './interaction'
import { PageState } from './state'
import type { Interaction } from './types'

function makeInteraction(
  component: string,
  event: string,
  actions: Interaction['then'],
): Interaction {
  return { when: { component, event }, then: actions }
}

describe('InteractionEngine', () => {
  it('dispatches matching interactions', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)
    const handler = vi.fn()
    engine.registerHandler('custom.action', handler)

    const interactions = [
      makeInteraction('chart', 'click', [
        { target: 'table', action: 'custom.action', params: { key: 'val' } },
      ]),
    ]

    engine.dispatch(interactions, 'chart', 'click', { x: 10 })
    expect(handler).toHaveBeenCalledTimes(1)
    expect(handler).toHaveBeenCalledWith(
      expect.objectContaining({ action: 'custom.action' }),
      { x: 10 },
    )
  })

  it('ignores non-matching interactions', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)
    const handler = vi.fn()
    engine.registerHandler('custom.action', handler)

    const interactions = [
      makeInteraction('chart', 'click', [{ target: 'table', action: 'custom.action' }]),
    ]

    engine.dispatch(interactions, 'other-component', 'click')
    expect(handler).not.toHaveBeenCalled()

    engine.dispatch(interactions, 'chart', 'hover')
    expect(handler).not.toHaveBeenCalled()
  })

  it('state.set handler sets state', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)

    const interactions = [
      makeInteraction('chart', 'pointSelected', [
        { target: '', action: 'state.set', params: { path: 'filters.month', value: 'July' } },
      ]),
    ]

    engine.dispatch(interactions, 'chart', 'pointSelected')
    expect(state.get('filters.month')).toBe('July')
  })

  it('state.set evaluates expressions in value', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)

    const interactions = [
      makeInteraction('chart', 'pointSelected', [
        { target: '', action: 'state.set', params: { path: 'selected', value: '${event.month}' } },
      ]),
    ]

    engine.dispatch(interactions, 'chart', 'pointSelected', { month: 'August' })
    expect(state.get('selected')).toBe('August')
  })

  it('state.toggle handler toggles boolean', () => {
    const state = new PageState()
    state.set('expanded', false)
    const engine = new InteractionEngine(state)

    const interactions = [
      makeInteraction('btn', 'click', [
        { target: '', action: 'state.toggle', params: { path: 'expanded' } },
      ]),
    ]

    engine.dispatch(interactions, 'btn', 'click')
    expect(state.get('expanded')).toBe(true)
  })

  it('skips action when condition is false', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)
    const handler = vi.fn()
    engine.registerHandler('test', handler)

    const interactions = [
      makeInteraction('btn', 'click', [
        { target: '', action: 'test', params: { condition: 'false' } },
      ]),
    ]

    engine.dispatch(interactions, 'btn', 'click')
    expect(handler).not.toHaveBeenCalled()
  })

  it('executes multiple actions in order', () => {
    const state = new PageState()
    const engine = new InteractionEngine(state)
    const order: string[] = []
    engine.registerHandler('first', () => order.push('first'))
    engine.registerHandler('second', () => order.push('second'))

    const interactions = [
      makeInteraction('btn', 'click', [
        { target: '', action: 'first' },
        { target: '', action: 'second' },
      ]),
    ]

    engine.dispatch(interactions, 'btn', 'click')
    expect(order).toEqual(['first', 'second'])
  })
})
