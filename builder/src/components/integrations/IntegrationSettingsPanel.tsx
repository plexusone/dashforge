import { useEffect, useState } from 'react'
import {
  X,
  Loader2,
  Trash2,
  Play,
  CheckCircle,
  XCircle,
  AlertCircle,
  MessageSquare,
  Mail,
  Globe,
  Slack,
} from 'lucide-react'
import { useIntegrationStore } from '../../stores/integration'
import type { TestConnectionResult } from '../../types/integration'
import clsx from 'clsx'

interface IntegrationSettingsPanelProps {
  isOpen: boolean
  onClose: () => void
}

const channelIcons: Record<string, React.ElementType> = {
  slack: Slack,
  email: Mail,
  whatsapp: MessageSquare,
  webhook: Globe,
}

const statusColors: Record<string, string> = {
  active: 'bg-green-100 text-green-700',
  inactive: 'bg-gray-100 text-gray-600',
  error: 'bg-red-100 text-red-700',
}

export function IntegrationSettingsPanel({ isOpen, onClose }: IntegrationSettingsPanelProps) {
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [testResult, setTestResult] = useState<TestConnectionResult | null>(null)
  const [isTesting, setIsTesting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  const {
    integrations,
    isLoadingIntegrations,
    fetchIntegrations,
    testIntegration,
    deleteIntegration,
    error,
    clearError,
  } = useIntegrationStore()

  useEffect(() => {
    if (isOpen) {
      fetchIntegrations()
    }
  }, [isOpen, fetchIntegrations])

  const selectedIntegration = integrations.find(i => i.id === selectedId)

  const handleTest = async () => {
    if (!selectedId) return
    setIsTesting(true)
    setTestResult(null)
    clearError()

    const result = await testIntegration(selectedId)
    setTestResult(result)
    setIsTesting(false)
  }

  const handleDelete = async () => {
    if (!selectedId) return
    if (!confirm('Are you sure you want to delete this integration?')) return

    setIsDeleting(true)
    const success = await deleteIntegration(selectedId)
    setIsDeleting(false)

    if (success) {
      setSelectedId(null)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Integrations</h2>
            <p className="text-sm text-gray-500">Manage your notification channels</p>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-hidden flex">
          {/* Left: Integration list */}
          <div className="w-64 border-r border-gray-200 flex flex-col">
            <div className="flex-1 overflow-y-auto">
              {isLoadingIntegrations ? (
                <div className="flex items-center justify-center h-full">
                  <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                </div>
              ) : integrations.length === 0 ? (
                <div className="text-center text-gray-500 py-8 px-4">
                  <p>No integrations configured</p>
                  <p className="text-xs mt-1">Install from the marketplace</p>
                </div>
              ) : (
                <div className="p-2">
                  {integrations.map((integration) => {
                    const Icon = channelIcons[integration.channelType] || Globe
                    return (
                      <button
                        key={integration.id}
                        onClick={() => {
                          setSelectedId(integration.id)
                          setTestResult(null)
                        }}
                        className={clsx(
                          'w-full p-3 rounded-lg text-left transition-colors mb-1',
                          selectedId === integration.id
                            ? 'bg-primary-50 border border-primary-200'
                            : 'hover:bg-gray-50'
                        )}
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
                            <Icon className="w-4 h-4 text-gray-600" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">
                              {integration.name}
                            </p>
                            <div className="flex items-center gap-2 mt-0.5">
                              <span className={clsx(
                                'text-xs px-1.5 py-0.5 rounded-full capitalize',
                                statusColors[integration.status]
                              )}>
                                {integration.status}
                              </span>
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

          {/* Right: Details */}
          <div className="flex-1 flex flex-col">
            {selectedIntegration ? (
              <>
                <div className="flex-1 overflow-y-auto p-6">
                  {/* Header */}
                  <div className="flex items-start justify-between mb-6">
                    <div className="flex items-center gap-3">
                      {(() => {
                        const Icon = channelIcons[selectedIntegration.channelType] || Globe
                        return (
                          <div className="w-12 h-12 rounded-lg bg-primary-100 flex items-center justify-center">
                            <Icon className="w-6 h-6 text-primary-600" />
                          </div>
                        )
                      })()}
                      <div>
                        <h3 className="font-semibold text-gray-900">{selectedIntegration.name}</h3>
                        <p className="text-sm text-gray-500">{selectedIntegration.slug}</p>
                      </div>
                    </div>
                    <span className={clsx(
                      'text-sm px-2 py-1 rounded-full capitalize',
                      statusColors[selectedIntegration.status]
                    )}>
                      {selectedIntegration.status}
                    </span>
                  </div>

                  {/* Details */}
                  <div className="space-y-4">
                    <div>
                      <label className="text-xs font-medium text-gray-500 uppercase">Channel Type</label>
                      <p className="text-sm text-gray-900 capitalize">{selectedIntegration.channelType}</p>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-gray-500 uppercase">Source</label>
                      <p className="text-sm text-gray-900 capitalize">{selectedIntegration.source}</p>
                    </div>

                    {selectedIntegration.statusMessage && (
                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase">Status Message</label>
                        <p className="text-sm text-red-600">{selectedIntegration.statusMessage}</p>
                      </div>
                    )}

                    {selectedIntegration.lastUsedAt && (
                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase">Last Used</label>
                        <p className="text-sm text-gray-900">
                          {new Date(selectedIntegration.lastUsedAt).toLocaleString()}
                        </p>
                      </div>
                    )}

                    <div>
                      <label className="text-xs font-medium text-gray-500 uppercase">Created</label>
                      <p className="text-sm text-gray-900">
                        {new Date(selectedIntegration.createdAt).toLocaleString()}
                      </p>
                    </div>

                    {/* Configuration (non-sensitive) */}
                    {selectedIntegration.config && Object.keys(selectedIntegration.config).length > 0 && (
                      <div>
                        <label className="text-xs font-medium text-gray-500 uppercase mb-2 block">Configuration</label>
                        <div className="bg-gray-50 rounded-lg p-3 text-sm">
                          {Object.entries(selectedIntegration.config).map(([key, value]) => (
                            <div key={key} className="flex justify-between py-1">
                              <span className="text-gray-600">{key}</span>
                              <span className="text-gray-900">{String(value)}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Test Result */}
                    {testResult && (
                      <div className={clsx(
                        'p-4 rounded-lg',
                        testResult.success ? 'bg-green-50' : 'bg-red-50'
                      )}>
                        <div className="flex items-center gap-2">
                          {testResult.success ? (
                            <CheckCircle className="w-5 h-5 text-green-600" />
                          ) : (
                            <XCircle className="w-5 h-5 text-red-600" />
                          )}
                          <span className={clsx(
                            'font-medium',
                            testResult.success ? 'text-green-700' : 'text-red-700'
                          )}>
                            {testResult.success ? 'Connection successful!' : 'Connection failed'}
                          </span>
                        </div>
                        {testResult.error && (
                          <p className="text-sm text-red-600 mt-2">{testResult.error}</p>
                        )}
                        <p className="text-xs text-gray-500 mt-2">
                          Completed in {testResult.durationMs}ms
                        </p>
                      </div>
                    )}

                    {error && (
                      <div className="p-3 bg-red-50 text-red-700 text-sm rounded-lg flex items-center gap-2">
                        <AlertCircle className="w-4 h-4" />
                        {error}
                      </div>
                    )}
                  </div>
                </div>

                {/* Actions */}
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

                  <button
                    onClick={handleTest}
                    disabled={isTesting}
                    className={clsx(
                      'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                      isTesting
                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                        : 'bg-primary-500 text-white hover:bg-primary-600'
                    )}
                  >
                    {isTesting ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Testing...
                      </>
                    ) : (
                      <>
                        <Play className="w-4 h-4" />
                        Test Connection
                      </>
                    )}
                  </button>
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-gray-500">
                <p>Select an integration to view details</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
