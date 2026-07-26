import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

const API_VERSION = 'ui.plexusone.dev/v1'
const KIND_PAGE = 'Page'

export interface PageMetadata {
  id: string
  name: string
  title?: string
  description?: string
  version?: string
  labels?: Record<string, string>
}

export type LayoutType =
  | 'responsive-grid'
  | 'stack'
  | 'split-pane'
  | 'tabs'
  | 'application-shell'

export interface LayoutConfig {
  columns?: number
  rows?: number
  gap?: string
  direction?: 'horizontal' | 'vertical'
  breakpoints?: Record<string, { columns: number; gap?: string }>
  sizes?: string[]
  resizable?: boolean
}

export interface LayoutRegion {
  name: string
  layout?: LayoutSpec
}

export interface LayoutSpec {
  type: LayoutType
  config?: LayoutConfig
  regions?: LayoutRegion[]
}

export interface Binding {
  source: string
  operation: string
  parameters?: Record<string, unknown>
  transform?: string
  default?: unknown
}

export interface Position {
  row?: number
  col?: number
  rowSpan?: number
  colSpan?: number
  order?: number
  region?: string
}

export interface VisibilityRule {
  condition?: string
  roles?: string[]
  capability?: string
}

export interface ComponentInstance {
  id: string
  type: string
  version?: string
  position?: Position
  properties?: Record<string, unknown>
  data?: Record<string, Binding>
  children?: ComponentInstance[]
  visibility?: VisibilityRule
  slot?: string
  style?: Record<string, string>
}

export interface InteractionTrigger {
  component: string
  event: string
}

export interface InteractionAction {
  target: string
  action: string
  value?: unknown
  condition?: string
  params?: Record<string, unknown>
}

export interface Interaction {
  when: InteractionTrigger
  then: InteractionAction[]
}

export interface NavItem {
  id: string
  label: string
  icon?: string
  target?: string
  children?: NavItem[]
}

export interface NavigationSpec {
  type: string
  items: NavItem[]
  position?: string
}

export interface ThemeRef {
  id: string
  variant?: string
  tokens?: Record<string, string>
}

export interface PageSpec {
  apiVersion: string
  kind: string
  metadata: PageMetadata
  profile?: string
  context?: Record<string, string>
  layout: LayoutSpec
  components: ComponentInstance[]
  interactions?: Interaction[]
  navigation?: NavigationSpec
  theme?: ThemeRef
}

interface HistoryState {
  past: PageSpec[]
  future: PageSpec[]
}

interface PageSpecState {
  page: PageSpec | null
  selectedComponentId: string | null
  isEditing: boolean
  isDirty: boolean
  history: HistoryState

  loadPageSpec: (page: PageSpec) => void
  newPageSpec: (profile?: string) => void
  exportPageSpec: () => PageSpec | null

  addComponent: (component: ComponentInstance) => void
  updateComponent: (id: string, updates: Partial<ComponentInstance>) => void
  removeComponent: (id: string) => void
  selectComponent: (id: string | null) => void

  addInteraction: (interaction: Interaction) => void
  removeInteraction: (index: number) => void

  updateLayout: (layout: Partial<LayoutSpec>) => void
  updateMetadata: (updates: Partial<PageMetadata>) => void

  undo: () => void
  redo: () => void
  canUndo: () => boolean
  canRedo: () => boolean

  markClean: () => void
}

const MAX_HISTORY = 50

function pushHistory(state: { page: PageSpec | null; history: HistoryState }): void {
  if (!state.page) return
  state.history.past.push(JSON.parse(JSON.stringify(state.page)))
  if (state.history.past.length > MAX_HISTORY) {
    state.history.past.shift()
  }
  state.history.future = []
}

export const usePageSpecStore = create<PageSpecState>()(
  devtools(
    immer((set, get) => ({
      page: null,
      selectedComponentId: null,
      isEditing: true,
      isDirty: false,
      history: { past: [], future: [] },

      loadPageSpec: (page) => set((state) => {
        state.page = page
        state.isDirty = false
        state.selectedComponentId = null
        state.history = { past: [], future: [] }
      }),

      newPageSpec: (profile) => set((state) => {
        state.page = {
          apiVersion: API_VERSION,
          kind: KIND_PAGE,
          metadata: {
            id: crypto.randomUUID(),
            name: 'untitled',
            title: 'Untitled Page',
          },
          profile: profile || 'dashboard',
          layout: {
            type: 'responsive-grid',
            config: { columns: 12, gap: '16px' },
          },
          components: [],
          interactions: [],
        }
        state.isDirty = false
        state.selectedComponentId = null
        state.history = { past: [], future: [] }
      }),

      exportPageSpec: () => {
        const { page } = get()
        if (!page) return null
        return JSON.parse(JSON.stringify(page))
      },

      addComponent: (component) => set((state) => {
        if (!state.page) return
        pushHistory(state)
        state.page.components.push(component)
        state.selectedComponentId = component.id
        state.isDirty = true
      }),

      updateComponent: (id, updates) => set((state) => {
        if (!state.page) return
        const component = state.page.components.find(c => c.id === id)
        if (!component) return
        pushHistory(state)
        Object.assign(component, updates)
        state.isDirty = true
      }),

      removeComponent: (id) => set((state) => {
        if (!state.page) return
        pushHistory(state)
        state.page.components = state.page.components.filter(c => c.id !== id)
        if (state.selectedComponentId === id) {
          state.selectedComponentId = null
        }
        state.isDirty = true
      }),

      selectComponent: (id) => set((state) => {
        state.selectedComponentId = id
      }),

      addInteraction: (interaction) => set((state) => {
        if (!state.page) return
        pushHistory(state)
        if (!state.page.interactions) {
          state.page.interactions = []
        }
        state.page.interactions.push(interaction)
        state.isDirty = true
      }),

      removeInteraction: (index) => set((state) => {
        if (!state.page?.interactions) return
        pushHistory(state)
        state.page.interactions.splice(index, 1)
        state.isDirty = true
      }),

      updateLayout: (layout) => set((state) => {
        if (!state.page) return
        pushHistory(state)
        Object.assign(state.page.layout, layout)
        state.isDirty = true
      }),

      updateMetadata: (updates) => set((state) => {
        if (!state.page) return
        pushHistory(state)
        Object.assign(state.page.metadata, updates)
        state.isDirty = true
      }),

      undo: () => set((state) => {
        if (state.history.past.length === 0 || !state.page) return
        const previous = state.history.past.pop()!
        state.history.future.push(JSON.parse(JSON.stringify(state.page)))
        state.page = previous
        state.isDirty = true
      }),

      redo: () => set((state) => {
        if (state.history.future.length === 0 || !state.page) return
        const next = state.history.future.pop()!
        state.history.past.push(JSON.parse(JSON.stringify(state.page)))
        state.page = next
        state.isDirty = true
      }),

      canUndo: () => get().history.past.length > 0,
      canRedo: () => get().history.future.length > 0,

      markClean: () => set((state) => {
        state.isDirty = false
      }),
    })),
    { name: 'pagespec-store' }
  )
)
