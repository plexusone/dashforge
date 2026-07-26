import type { ComponentType } from 'react'
import type { ComponentInstance } from './types'

export interface ComponentProps {
  instance: ComponentInstance
  children?: React.ReactNode
}

export type UIForgeComponent = ComponentType<ComponentProps>

const components = new Map<string, UIForgeComponent>()

export function registerComponent(type: string, component: UIForgeComponent): void {
  components.set(type, component)
}

export function getComponent(type: string): UIForgeComponent | undefined {
  return components.get(type)
}

export function hasComponent(type: string): boolean {
  return components.has(type)
}

export function listComponents(): string[] {
  return Array.from(components.keys())
}

export function clearRegistry(): void {
  components.clear()
}
