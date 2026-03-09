import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import type {
  Integration,
  Alert,
  AlertEvent,
  IntegrationDefinition,
  ChannelInfo,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  CreateAlertRequest,
  UpdateAlertRequest,
  InstallIntegrationRequest,
  TestConnectionResult,
} from '../types/integration'
import * as api from '../api/integration'

interface IntegrationState {
  // Data
  integrations: Integration[]
  alerts: Alert[]
  alertEvents: Record<number, AlertEvent[]> // keyed by alert ID
  marketplaceIntegrations: IntegrationDefinition[]
  channels: ChannelInfo[]

  // Loading states
  isLoadingIntegrations: boolean
  isLoadingAlerts: boolean
  isLoadingMarketplace: boolean

  // Selection
  selectedIntegrationId: number | null
  selectedAlertId: number | null

  // Error state
  error: string | null

  // Integration actions
  fetchIntegrations: () => Promise<void>
  fetchIntegration: (id: number) => Promise<Integration | null>
  createIntegration: (request: CreateIntegrationRequest) => Promise<Integration | null>
  updateIntegration: (id: number, request: UpdateIntegrationRequest) => Promise<Integration | null>
  deleteIntegration: (id: number) => Promise<boolean>
  testIntegration: (id: number) => Promise<TestConnectionResult | null>
  selectIntegration: (id: number | null) => void

  // Alert actions
  fetchAlerts: () => Promise<void>
  fetchAlert: (id: number) => Promise<Alert | null>
  createAlert: (request: CreateAlertRequest) => Promise<Alert | null>
  updateAlert: (id: number, request: UpdateAlertRequest) => Promise<Alert | null>
  deleteAlert: (id: number) => Promise<boolean>
  enableAlert: (id: number) => Promise<Alert | null>
  disableAlert: (id: number) => Promise<Alert | null>
  fetchAlertEvents: (alertId: number, limit?: number) => Promise<void>
  selectAlert: (id: number | null) => void

  // Marketplace actions
  fetchMarketplace: (category?: string) => Promise<void>
  installIntegration: (marketplaceSlug: string, request: InstallIntegrationRequest) => Promise<Integration | null>

  // Channel actions
  fetchChannels: () => Promise<void>

  // Utility
  clearError: () => void
}

export const useIntegrationStore = create<IntegrationState>()(
  devtools(
    immer((set, get) => ({
      // Initial state
      integrations: [],
      alerts: [],
      alertEvents: {},
      marketplaceIntegrations: [],
      channels: [],
      isLoadingIntegrations: false,
      isLoadingAlerts: false,
      isLoadingMarketplace: false,
      selectedIntegrationId: null,
      selectedAlertId: null,
      error: null,

      // Integration actions

      fetchIntegrations: async () => {
        set((state) => {
          state.isLoadingIntegrations = true
          state.error = null
        })
        try {
          const response = await api.listIntegrations()
          set((state) => {
            state.integrations = response.integrations
            state.isLoadingIntegrations = false
          })
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch integrations'
            state.isLoadingIntegrations = false
          })
        }
      },

      fetchIntegration: async (id) => {
        try {
          const integration = await api.getIntegration(id)
          set((state) => {
            const index = state.integrations.findIndex(i => i.id === id)
            if (index >= 0) {
              state.integrations[index] = integration
            } else {
              state.integrations.push(integration)
            }
          })
          return integration
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch integration'
          })
          return null
        }
      },

      createIntegration: async (request) => {
        try {
          const integration = await api.createIntegration(request)
          set((state) => {
            state.integrations.push(integration)
          })
          return integration
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to create integration'
          })
          return null
        }
      },

      updateIntegration: async (id, request) => {
        try {
          const integration = await api.updateIntegration(id, request)
          set((state) => {
            const index = state.integrations.findIndex(i => i.id === id)
            if (index >= 0) {
              state.integrations[index] = integration
            }
          })
          return integration
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to update integration'
          })
          return null
        }
      },

      deleteIntegration: async (id) => {
        try {
          await api.deleteIntegration(id)
          set((state) => {
            state.integrations = state.integrations.filter(i => i.id !== id)
            if (state.selectedIntegrationId === id) {
              state.selectedIntegrationId = null
            }
          })
          return true
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to delete integration'
          })
          return false
        }
      },

      testIntegration: async (id) => {
        try {
          const result = await api.testIntegration(id)
          // Refresh the integration to get updated status
          await get().fetchIntegration(id)
          return result
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to test integration'
          })
          return null
        }
      },

      selectIntegration: (id) => {
        set((state) => {
          state.selectedIntegrationId = id
        })
      },

      // Alert actions

      fetchAlerts: async () => {
        set((state) => {
          state.isLoadingAlerts = true
          state.error = null
        })
        try {
          const response = await api.listAlerts()
          set((state) => {
            state.alerts = response.alerts
            state.isLoadingAlerts = false
          })
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch alerts'
            state.isLoadingAlerts = false
          })
        }
      },

      fetchAlert: async (id) => {
        try {
          const alert = await api.getAlert(id)
          set((state) => {
            const index = state.alerts.findIndex(a => a.id === id)
            if (index >= 0) {
              state.alerts[index] = alert
            } else {
              state.alerts.push(alert)
            }
          })
          return alert
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch alert'
          })
          return null
        }
      },

      createAlert: async (request) => {
        try {
          const alert = await api.createAlert(request)
          set((state) => {
            state.alerts.push(alert)
          })
          return alert
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to create alert'
          })
          return null
        }
      },

      updateAlert: async (id, request) => {
        try {
          const alert = await api.updateAlert(id, request)
          set((state) => {
            const index = state.alerts.findIndex(a => a.id === id)
            if (index >= 0) {
              state.alerts[index] = alert
            }
          })
          return alert
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to update alert'
          })
          return null
        }
      },

      deleteAlert: async (id) => {
        try {
          await api.deleteAlert(id)
          set((state) => {
            state.alerts = state.alerts.filter(a => a.id !== id)
            if (state.selectedAlertId === id) {
              state.selectedAlertId = null
            }
            delete state.alertEvents[id]
          })
          return true
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to delete alert'
          })
          return false
        }
      },

      enableAlert: async (id) => {
        try {
          const alert = await api.enableAlert(id)
          set((state) => {
            const index = state.alerts.findIndex(a => a.id === id)
            if (index >= 0) {
              state.alerts[index] = alert
            }
          })
          return alert
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to enable alert'
          })
          return null
        }
      },

      disableAlert: async (id) => {
        try {
          const alert = await api.disableAlert(id)
          set((state) => {
            const index = state.alerts.findIndex(a => a.id === id)
            if (index >= 0) {
              state.alerts[index] = alert
            }
          })
          return alert
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to disable alert'
          })
          return null
        }
      },

      fetchAlertEvents: async (alertId, limit) => {
        try {
          const response = await api.getAlertEvents(alertId, limit)
          set((state) => {
            state.alertEvents[alertId] = response.events
          })
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch alert events'
          })
        }
      },

      selectAlert: (id) => {
        set((state) => {
          state.selectedAlertId = id
        })
      },

      // Marketplace actions

      fetchMarketplace: async (category) => {
        set((state) => {
          state.isLoadingMarketplace = true
          state.error = null
        })
        try {
          const response = await api.listMarketplaceIntegrations(category)
          set((state) => {
            state.marketplaceIntegrations = response.integrations
            state.isLoadingMarketplace = false
          })
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch marketplace'
            state.isLoadingMarketplace = false
          })
        }
      },

      installIntegration: async (marketplaceSlug, request) => {
        try {
          const integration = await api.installMarketplaceIntegration(marketplaceSlug, request)
          set((state) => {
            state.integrations.push(integration)
          })
          return integration
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to install integration'
          })
          return null
        }
      },

      // Channel actions

      fetchChannels: async () => {
        try {
          const response = await api.listChannels()
          set((state) => {
            state.channels = response.channels
          })
        } catch (err) {
          set((state) => {
            state.error = err instanceof Error ? err.message : 'Failed to fetch channels'
          })
        }
      },

      // Utility

      clearError: () => {
        set((state) => {
          state.error = null
        })
      },
    })),
    { name: 'integration-store' }
  )
)

// Selectors
export const selectIntegrations = (state: IntegrationState) => state.integrations
export const selectAlerts = (state: IntegrationState) => state.alerts
export const selectMarketplace = (state: IntegrationState) => state.marketplaceIntegrations
export const selectChannels = (state: IntegrationState) => state.channels
export const selectSelectedIntegration = (state: IntegrationState) =>
  state.integrations.find(i => i.id === state.selectedIntegrationId)
export const selectSelectedAlert = (state: IntegrationState) =>
  state.alerts.find(a => a.id === state.selectedAlertId)
export const selectAlertEvents = (alertId: number) => (state: IntegrationState) =>
  state.alertEvents[alertId] || []
