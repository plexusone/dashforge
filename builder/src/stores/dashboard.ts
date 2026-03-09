import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import type { Dashboard, Widget, DataSource, Variable, Position } from '../types/dashboard'

interface HistoryState {
  past: Dashboard[]
  future: Dashboard[]
}

interface DashboardState {
  // Current dashboard
  dashboard: Dashboard

  // Selection state
  selectedWidgetId: string | null

  // Edit mode
  isEditing: boolean

  // Dirty state
  isDirty: boolean

  // History for undo/redo
  history: HistoryState

  // Actions
  setDashboard: (dashboard: Dashboard) => void
  updateDashboard: (updates: Partial<Dashboard>) => void

  // Widget actions
  addWidget: (widget: Widget) => void
  updateWidget: (widgetId: string, updates: Partial<Widget>) => void
  removeWidget: (widgetId: string) => void
  moveWidget: (widgetId: string, position: Partial<Position>) => void
  duplicateWidget: (widgetId: string) => void

  // Selection actions
  selectWidget: (widgetId: string | null) => void

  // DataSource actions
  addDataSource: (dataSource: DataSource) => void
  updateDataSource: (dataSourceId: string, updates: Partial<DataSource>) => void
  removeDataSource: (dataSourceId: string) => void

  // Variable actions
  addVariable: (variable: Variable) => void
  updateVariable: (variableId: string, updates: Partial<Variable>) => void
  removeVariable: (variableId: string) => void

  // Edit mode
  setEditMode: (isEditing: boolean) => void

  // History actions
  undo: () => void
  redo: () => void
  canUndo: () => boolean
  canRedo: () => boolean

  // Persistence
  markClean: () => void
  exportDashboard: () => Dashboard
}

const createEmptyDashboard = (): Dashboard => ({
  id: crypto.randomUUID(),
  title: 'Untitled Dashboard',
  description: '',
  layout: {
    type: 'grid',
    columns: 12,
    rowHeight: 80,
    gap: 8,
    padding: 16
  },
  theme: {
    name: 'default',
    colors: {
      primary: '#0ea5e9',
      secondary: '#64748b',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#0f172a',
      border: '#e2e8f0'
    }
  },
  variables: [],
  dataSources: [],
  widgets: []
})

const MAX_HISTORY = 50

export const useDashboardStore = create<DashboardState>()(
  devtools(
    immer((set, get) => ({
      dashboard: createEmptyDashboard(),
      selectedWidgetId: null,
      isEditing: true,
      isDirty: false,
      history: {
        past: [],
        future: []
      },

      setDashboard: (dashboard) => set((state) => {
        state.dashboard = dashboard
        state.isDirty = false
        state.history = { past: [], future: [] }
      }),

      updateDashboard: (updates) => set((state) => {
        // Save current state to history
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        // Apply updates
        Object.assign(state.dashboard, updates)
        state.isDirty = true
      }),

      addWidget: (widget) => set((state) => {
        // Save to history
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        state.dashboard.widgets.push(widget)
        state.selectedWidgetId = widget.id
        state.isDirty = true
      }),

      updateWidget: (widgetId, updates) => set((state) => {
        const widget = state.dashboard.widgets.find(w => w.id === widgetId)
        if (widget) {
          // Save to history
          state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
          if (state.history.past.length > MAX_HISTORY) {
            state.history.past.shift()
          }
          state.history.future = []

          Object.assign(widget, updates)
          state.isDirty = true
        }
      }),

      removeWidget: (widgetId) => set((state) => {
        // Save to history
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        state.dashboard.widgets = state.dashboard.widgets.filter(w => w.id !== widgetId)
        if (state.selectedWidgetId === widgetId) {
          state.selectedWidgetId = null
        }
        state.isDirty = true
      }),

      moveWidget: (widgetId, position) => set((state) => {
        const widget = state.dashboard.widgets.find(w => w.id === widgetId)
        if (widget) {
          Object.assign(widget.position, position)
          state.isDirty = true
        }
      }),

      duplicateWidget: (widgetId) => set((state) => {
        const widget = state.dashboard.widgets.find(w => w.id === widgetId)
        if (widget) {
          // Save to history
          state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
          if (state.history.past.length > MAX_HISTORY) {
            state.history.past.shift()
          }
          state.history.future = []

          const newWidget: Widget = {
            ...JSON.parse(JSON.stringify(widget)),
            id: crypto.randomUUID(),
            position: {
              ...widget.position,
              y: widget.position.y + widget.position.h
            }
          }
          state.dashboard.widgets.push(newWidget)
          state.selectedWidgetId = newWidget.id
          state.isDirty = true
        }
      }),

      selectWidget: (widgetId) => set((state) => {
        state.selectedWidgetId = widgetId
      }),

      addDataSource: (dataSource) => set((state) => {
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        state.dashboard.dataSources.push(dataSource)
        state.isDirty = true
      }),

      updateDataSource: (dataSourceId, updates) => set((state) => {
        const dataSource = state.dashboard.dataSources.find(d => d.id === dataSourceId)
        if (dataSource) {
          state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
          if (state.history.past.length > MAX_HISTORY) {
            state.history.past.shift()
          }
          state.history.future = []

          Object.assign(dataSource, updates)
          state.isDirty = true
        }
      }),

      removeDataSource: (dataSourceId) => set((state) => {
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        state.dashboard.dataSources = state.dashboard.dataSources.filter(d => d.id !== dataSourceId)
        state.isDirty = true
      }),

      addVariable: (variable) => set((state) => {
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        if (!state.dashboard.variables) {
          state.dashboard.variables = []
        }
        state.dashboard.variables.push(variable)
        state.isDirty = true
      }),

      updateVariable: (variableId, updates) => set((state) => {
        const variable = state.dashboard.variables?.find(v => v.id === variableId)
        if (variable) {
          state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
          if (state.history.past.length > MAX_HISTORY) {
            state.history.past.shift()
          }
          state.history.future = []

          Object.assign(variable, updates)
          state.isDirty = true
        }
      }),

      removeVariable: (variableId) => set((state) => {
        state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
        if (state.history.past.length > MAX_HISTORY) {
          state.history.past.shift()
        }
        state.history.future = []

        if (state.dashboard.variables) {
          state.dashboard.variables = state.dashboard.variables.filter(v => v.id !== variableId)
        }
        state.isDirty = true
      }),

      setEditMode: (isEditing) => set((state) => {
        state.isEditing = isEditing
      }),

      undo: () => set((state) => {
        if (state.history.past.length > 0) {
          const previous = state.history.past.pop()!
          state.history.future.push(JSON.parse(JSON.stringify(state.dashboard)))
          state.dashboard = previous
          state.isDirty = true
        }
      }),

      redo: () => set((state) => {
        if (state.history.future.length > 0) {
          const next = state.history.future.pop()!
          state.history.past.push(JSON.parse(JSON.stringify(state.dashboard)))
          state.dashboard = next
          state.isDirty = true
        }
      }),

      canUndo: () => get().history.past.length > 0,
      canRedo: () => get().history.future.length > 0,

      markClean: () => set((state) => {
        state.isDirty = false
      }),

      exportDashboard: () => JSON.parse(JSON.stringify(get().dashboard))
    })),
    { name: 'dashboard-store' }
  )
)
