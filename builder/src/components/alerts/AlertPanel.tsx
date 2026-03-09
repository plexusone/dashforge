import { useEffect, useState } from 'react'
import {
  X,
  Loader2,
  Plus,
  Trash2,
  Bell,
  BellOff,
  Clock,
  AlertTriangle,
  CheckCircle,
  XCircle,
  History,
} from 'lucide-react'
import { useIntegrationStore } from '../../stores/integration'
import { TriggerForm } from './TriggerForms'
import type { CreateAlertRequest, TriggerType } from '../../types/integration'
import clsx from 'clsx'

interface AlertPanelProps {
  isOpen: boolean
  onClose: () => void
}

const triggerTypeLabels: Record<TriggerType, string> = {
  threshold: 'Threshold',
  schedule: 'Schedule',
  data_change: 'Data Change',
}

const triggerTypeIcons: Record<TriggerType, React.ElementType> = {
  threshold: AlertTriangle,
  schedule: Clock,
  data_change: History,
}

export function AlertPanel({ isOpen, onClose }: AlertPanelProps) {
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [newAlertData, setNewAlertData] = useState<Partial<CreateAlertRequest>>({
    name: '',
    slug: '',
    triggerType: 'threshold',
    triggerConfig: {},
    cooldownSeconds: 300,
    enabled: true,
    channelIds: [],
  })
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showHistory, setShowHistory] = useState(false)

  const {
    alerts,
    alertEvents,
    integrations,
    isLoadingAlerts,
    fetchAlerts,
    fetchIntegrations,
    fetchAlertEvents,
    createAlert,
    deleteAlert,
    enableAlert,
    disableAlert,
    error,
    clearError,
  } = useIntegrationStore()

  useEffect(() => {
    if (isOpen) {
      fetchAlerts()
      fetchIntegrations()
    }
  }, [isOpen, fetchAlerts, fetchIntegrations])

  useEffect(() => {
    if (selectedId && showHistory) {
      fetchAlertEvents(selectedId, 20)
    }
  }, [selectedId, showHistory, fetchAlertEvents])

  const selectedAlert = alerts.find(a => a.id === selectedId)
  const selectedAlertEvents = selectedId ? alertEvents[selectedId] || [] : []

  const handleCreate = async () => {
    if (!newAlertData.name || !newAlertData.slug) return

    setIsSaving(true)
    clearError()

    const result = await createAlert(newAlertData as CreateAlertRequest)
    setIsSaving(false)

    if (result) {
      setIsCreating(false)
      setNewAlertData({
        name: '',
        slug: '',
        triggerType: 'threshold',
        triggerConfig: {},
        cooldownSeconds: 300,
        enabled: true,
        channelIds: [],
      })
      setSelectedId(result.id)
    }
  }

  const handleDelete = async () => {
    if (!selectedId) return
    if (!confirm('Are you sure you want to delete this alert?')) return

    setIsDeleting(true)
    const success = await deleteAlert(selectedId)
    setIsDeleting(false)

    if (success) {
      setSelectedId(null)
    }
  }

  const handleToggleEnabled = async () => {
    if (!selectedAlert) return
    if (selectedAlert.enabled) {
      await disableAlert(selectedAlert.id)
    } else {
      await enableAlert(selectedAlert.id)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Alerts</h2>
            <p className="text-sm text-gray-500">Configure alerts for your dashboards</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                setIsCreating(true)
                setSelectedId(null)
              }}
              className="flex items-center gap-2 px-3 py-2 text-sm bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              <Plus className="w-4 h-4" />
              New Alert
            </button>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <X className="w-5 h-5 text-gray-500" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-hidden flex">
          {/* Left: Alert list */}
          <div className="w-72 border-r border-gray-200 flex flex-col">
            <div className="flex-1 overflow-y-auto">
              {isLoadingAlerts ? (
                <div className="flex items-center justify-center h-full">
                  <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                </div>
              ) : alerts.length === 0 ? (
                <div className="text-center text-gray-500 py-8 px-4">
                  <p>No alerts configured</p>
                  <p className="text-xs mt-1">Create your first alert</p>
                </div>
              ) : (
                <div className="p-2">
                  {alerts.map((alert) => {
                    const Icon = triggerTypeIcons[alert.triggerType]
                    return (
                      <button
                        key={alert.id}
                        onClick={() => {
                          setSelectedId(alert.id)
                          setIsCreating(false)
                          setShowHistory(false)
                        }}
                        className={clsx(
                          'w-full p-3 rounded-lg text-left transition-colors mb-1',
                          selectedId === alert.id
                            ? 'bg-primary-50 border border-primary-200'
                            : 'hover:bg-gray-50'
                        )}
                      >
                        <div className="flex items-start gap-3">
                          <div className={clsx(
                            'w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0',
                            alert.enabled ? 'bg-green-100' : 'bg-gray-100'
                          )}>
                            <Icon className={clsx(
                              'w-4 h-4',
                              alert.enabled ? 'text-green-600' : 'text-gray-400'
                            )} />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">
                              {alert.name}
                            </p>
                            <div className="flex items-center gap-2 mt-0.5">
                              <span className="text-xs text-gray-500">
                                {triggerTypeLabels[alert.triggerType]}
                              </span>
                              {!alert.enabled && (
                                <span className="text-xs px-1.5 py-0.5 bg-gray-100 text-gray-500 rounded-full">
                                  Disabled
                                </span>
                              )}
                            </div>
                          </div>
                        </div>
                      </button>
                    )
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Right: Details/Create */}
          <div className="flex-1 flex flex-col">
            {isCreating ? (
              // Create new alert form
              <>
                <div className="flex-1 overflow-y-auto p-6">
                  <h3 className="font-semibold text-gray-900 mb-4">Create New Alert</h3>

                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Name *
                      </label>
                      <input
                        type="text"
                        value={newAlertData.name}
                        onChange={(e) => setNewAlertData({ ...newAlertData, name: e.target.value })}
                        placeholder="My Alert"
                        className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Slug *
                      </label>
                      <input
                        type="text"
                        value={newAlertData.slug}
                        onChange={(e) => setNewAlertData({ ...newAlertData, slug: e.target.value })}
                        placeholder="my-alert"
                        className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Trigger Type
                      </label>
                      <select
                        value={newAlertData.triggerType}
                        onChange={(e) => setNewAlertData({
                          ...newAlertData,
                          triggerType: e.target.value as TriggerType,
                          triggerConfig: {},
                        })}
                        className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      >
                        <option value="threshold">Threshold</option>
                        <option value="schedule">Schedule</option>
                        <option value="data_change">Data Change</option>
                      </select>
                    </div>

                    <TriggerForm
                      triggerType={newAlertData.triggerType!}
                      config={newAlertData.triggerConfig || {}}
                      onChange={(config) => setNewAlertData({ ...newAlertData, triggerConfig: config })}
                    />

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Cooldown (seconds)
                      </label>
                      <input
                        type="number"
                        value={newAlertData.cooldownSeconds}
                        onChange={(e) => setNewAlertData({ ...newAlertData, cooldownSeconds: parseInt(e.target.value) || 300 })}
                        min={60}
                        className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      />
                      <p className="text-xs text-gray-500 mt-1">Minimum time between alerts</p>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Notification Channels
                      </label>
                      <div className="space-y-2">
                        {integrations.map((integration) => (
                          <label key={integration.id} className="flex items-center gap-2">
                            <input
                              type="checkbox"
                              checked={newAlertData.channelIds?.includes(integration.id)}
                              onChange={(e) => {
                                const ids = newAlertData.channelIds || []
                                if (e.target.checked) {
                                  setNewAlertData({ ...newAlertData, channelIds: [...ids, integration.id] })
                                } else {
                                  setNewAlertData({ ...newAlertData, channelIds: ids.filter(id => id !== integration.id) })
                                }
                              }}
                              className="rounded border-gray-300 text-primary-500 focus:ring-primary-500"
                            />
                            <span className="text-sm text-gray-700">{integration.name}</span>
                            <span className="text-xs text-gray-500 capitalize">({integration.channelType})</span>
                          </label>
                        ))}
                        {integrations.length === 0 && (
                          <p className="text-sm text-gray-500">No integrations configured</p>
                        )}
                      </div>
                    </div>

                    {error && (
                      <div className="p-3 bg-red-50 text-red-700 text-sm rounded-lg">
                        {error}
                      </div>
                    )}
                  </div>
                </div>

                <div className="p-4 border-t border-gray-200 flex items-center justify-end gap-3">
                  <button
                    onClick={() => setIsCreating(false)}
                    className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleCreate}
                    disabled={!newAlertData.name || !newAlertData.slug || isSaving}
                    className={clsx(
                      'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                      !newAlertData.name || !newAlertData.slug || isSaving
                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                        : 'bg-primary-500 text-white hover:bg-primary-600'
                    )}
                  >
                    {isSaving ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Creating...
                      </>
                    ) : (
                      'Create Alert'
                    )}
                  </button>
                </div>
              </>
            ) : selectedAlert ? (
              // Alert details
              <>
                <div className="flex-1 overflow-y-auto p-6">
                  {/* Header */}
                  <div className="flex items-start justify-between mb-6">
                    <div>
                      <h3 className="font-semibold text-gray-900">{selectedAlert.name}</h3>
                      <p className="text-sm text-gray-500">{selectedAlert.slug}</p>
                    </div>
                    <button
                      onClick={handleToggleEnabled}
                      className={clsx(
                        'flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors',
                        selectedAlert.enabled
                          ? 'bg-green-100 text-green-700 hover:bg-green-200'
                          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                      )}
                    >
                      {selectedAlert.enabled ? (
                        <>
                          <Bell className="w-4 h-4" />
                          Enabled
                        </>
                      ) : (
                        <>
                          <BellOff className="w-4 h-4" />
                          Disabled
                        </>
                      )}
                    </button>
                  </div>

                  {/* Tabs */}
                  <div className="flex gap-4 mb-4 border-b border-gray-200">
                    <button
                      onClick={() => setShowHistory(false)}
                      className={clsx(
                        'pb-2 text-sm font-medium transition-colors',
                        !showHistory
                          ? 'text-primary-600 border-b-2 border-primary-500'
                          : 'text-gray-500 hover:text-gray-700'
                      )}
                    >
                      Details
                    </button>
                    <button
                      onClick={() => setShowHistory(true)}
                      className={clsx(
                        'pb-2 text-sm font-medium transition-colors',
                        showHistory
                          ? 'text-primary-600 border-b-2 border-primary-500'
                          : 'text-gray-500 hover:text-gray-700'
                      )}
                    >
                      History
                    </button>
                  </div>

                  {!showHistory ? (
                    // Details tab
                    <div className="space-y-4">
                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase">Trigger Type</label>
                        <p className="text-sm text-gray-900">{triggerTypeLabels[selectedAlert.triggerType]}</p>
                      </div>

                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase">Cooldown</label>
                        <p className="text-sm text-gray-900">{selectedAlert.cooldownSeconds} seconds</p>
                      </div>

                      {selectedAlert.lastTriggeredAt && (
                        <div>
                          <label className="text-xs font-medium text-gray-500 uppercase">Last Triggered</label>
                          <p className="text-sm text-gray-900">
                            {new Date(selectedAlert.lastTriggeredAt).toLocaleString()}
                          </p>
                        </div>
                      )}

                      {selectedAlert.lastError && (
                        <div>
                          <label className="text-xs font-medium text-gray-500 uppercase">Last Error</label>
                          <p className="text-sm text-red-600">{selectedAlert.lastError}</p>
                        </div>
                      )}

                      {selectedAlert.channels && selectedAlert.channels.length > 0 && (
                        <div>
                          <label className="text-xs font-medium text-gray-500 uppercase mb-2 block">Channels</label>
                          <div className="flex flex-wrap gap-2">
                            {selectedAlert.channels.map((channel) => (
                              <span
                                key={channel.id}
                                className="px-2 py-1 bg-gray-100 text-gray-700 text-sm rounded-full"
                              >
                                {channel.name}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}

                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase mb-2 block">Trigger Configuration</label>
                        <pre className="bg-gray-50 rounded-lg p-3 text-xs text-gray-700 overflow-x-auto">
                          {JSON.stringify(selectedAlert.triggerConfig, null, 2)}
                        </pre>
                      </div>
                    </div>
                  ) : (
                    // History tab
                    <div className="space-y-3">
                      {selectedAlertEvents.length === 0 ? (
                        <p className="text-sm text-gray-500 text-center py-4">No events recorded</p>
                      ) : (
                        selectedAlertEvents.map((event) => (
                          <div
                            key={event.id}
                            className={clsx(
                              'p-3 rounded-lg border',
                              event.eventType === 'triggered' && 'bg-yellow-50 border-yellow-200',
                              event.eventType === 'resolved' && 'bg-green-50 border-green-200',
                              event.eventType === 'error' && 'bg-red-50 border-red-200',
                              !['triggered', 'resolved', 'error'].includes(event.eventType) && 'bg-gray-50 border-gray-200'
                            )}
                          >
                            <div className="flex items-center gap-2 mb-1">
                              {event.eventType === 'triggered' && <AlertTriangle className="w-4 h-4 text-yellow-600" />}
                              {event.eventType === 'resolved' && <CheckCircle className="w-4 h-4 text-green-600" />}
                              {event.eventType === 'error' && <XCircle className="w-4 h-4 text-red-600" />}
                              <span className="text-sm font-medium capitalize">{event.eventType.replace('_', ' ')}</span>
                              <span className="text-xs text-gray-500 ml-auto">
                                {new Date(event.createdAt).toLocaleString()}
                              </span>
                            </div>
                            {event.channelsNotified && event.channelsNotified.length > 0 && (
                              <p className="text-xs text-gray-600">
                                Notified: {event.channelsNotified.join(', ')}
                                {event.channelsFailed > 0 && ` (${event.channelsFailed} failed)`}
                              </p>
                            )}
                            {event.errorMessage && (
                              <p className="text-xs text-red-600 mt-1">{event.errorMessage}</p>
                            )}
                          </div>
                        ))
                      )}
                    </div>
                  )}
                </div>

                <div className="p-4 border-t border-gray-200 flex items-center justify-between">
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    {isDeleting ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      <Trash2 className="w-4 h-4" />
                    )}
                    Delete
                  </button>
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-gray-500">
                <p>Select an alert or create a new one</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
