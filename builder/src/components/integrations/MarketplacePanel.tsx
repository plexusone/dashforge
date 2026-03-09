import { useEffect, useState } from 'react'
import {
  X,
  Search,
  Loader2,
  MessageSquare,
  Mail,
  Globe,
  Slack,
  CheckCircle,
} from 'lucide-react'
import { useIntegrationStore } from '../../stores/integration'
import type { IntegrationDefinition, InstallIntegrationRequest } from '../../types/integration'
import clsx from 'clsx'

interface MarketplacePanelProps {
  isOpen: boolean
  onClose: () => void
}

const channelIcons: Record<string, React.ElementType> = {
  slack: Slack,
  email: Mail,
  whatsapp: MessageSquare,
  webhook: Globe,
}

export function MarketplacePanel({ isOpen, onClose }: MarketplacePanelProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationDefinition | null>(null)
  const [installSlug, setInstallSlug] = useState('')
  const [installConfig, setInstallConfig] = useState<Record<string, string>>({})
  const [isInstalling, setIsInstalling] = useState(false)

  const {
    marketplaceIntegrations,
    isLoadingMarketplace,
    fetchMarketplace,
    installIntegration,
    error,
    clearError,
  } = useIntegrationStore()

  useEffect(() => {
    if (isOpen) {
      fetchMarketplace()
    }
  }, [isOpen, fetchMarketplace])

  const filteredIntegrations = marketplaceIntegrations.filter((integration) => {
    const matchesSearch = searchQuery === '' ||
      integration.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      integration.description.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesCategory = selectedCategory === null || integration.category === selectedCategory
    return matchesSearch && matchesCategory
  })

  const categories = [...new Set(marketplaceIntegrations.map(i => i.category))]

  const handleInstall = async () => {
    if (!selectedIntegration || !installSlug) return

    setIsInstalling(true)
    clearError()

    const request: InstallIntegrationRequest = {
      slug: installSlug,
      name: selectedIntegration.name,
      config: {},
      credentials: {},
    }

    // Separate config and credentials based on schema
    Object.entries(installConfig).forEach(([key, value]) => {
      const property = selectedIntegration.configSchema.properties[key]
      if (property?.secret) {
        request.credentials![key] = value
      } else {
        request.config![key] = value
      }
    })

    const result = await installIntegration(selectedIntegration.slug, request)
    setIsInstalling(false)

    if (result) {
      setSelectedIntegration(null)
      setInstallSlug('')
      setInstallConfig({})
      onClose()
    }
  }

  if (!isOpen) return null

  const Icon = selectedIntegration ? channelIcons[selectedIntegration.channelType] || Globe : Globe

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Integration Marketplace</h2>
            <p className="text-sm text-gray-500">Browse and install notification integrations</p>
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
          {/* Left: Browse */}
          <div className="flex-1 flex flex-col border-r border-gray-200">
            {/* Search */}
            <div className="p-4 border-b border-gray-100">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search integrations..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
              </div>

              {/* Categories */}
              <div className="flex gap-2 mt-3">
                <button
                  onClick={() => setSelectedCategory(null)}
                  className={clsx(
                    'px-3 py-1 text-xs rounded-full transition-colors',
                    selectedCategory === null
                      ? 'bg-primary-500 text-white'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  )}
                >
                  All
                </button>
                {categories.map((category) => (
                  <button
                    key={category}
                    onClick={() => setSelectedCategory(category)}
                    className={clsx(
                      'px-3 py-1 text-xs rounded-full transition-colors capitalize',
                      selectedCategory === category
                        ? 'bg-primary-500 text-white'
                        : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    )}
                  >
                    {category}
                  </button>
                ))}
              </div>
            </div>

            {/* Integrations list */}
            <div className="flex-1 overflow-y-auto p-4">
              {isLoadingMarketplace ? (
                <div className="flex items-center justify-center h-full">
                  <Loader2 className="w-6 h-6 animate-spin text-gray-400" />
                </div>
              ) : filteredIntegrations.length === 0 ? (
                <div className="text-center text-gray-500 py-8">
                  No integrations found
                </div>
              ) : (
                <div className="grid gap-3">
                  {filteredIntegrations.map((integration) => {
                    const ChannelIcon = channelIcons[integration.channelType] || Globe
                    return (
                      <button
                        key={integration.slug}
                        onClick={() => {
                          setSelectedIntegration(integration)
                          setInstallSlug(integration.slug + '-' + Date.now())
                          setInstallConfig({})
                        }}
                        className={clsx(
                          'w-full p-4 rounded-lg border text-left transition-colors',
                          selectedIntegration?.slug === integration.slug
                            ? 'border-primary-500 bg-primary-50'
                            : 'border-gray-200 hover:border-gray-300'
                        )}
                      >
                        <div className="flex items-start gap-3">
                          <div className="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center flex-shrink-0">
                            <ChannelIcon className="w-5 h-5 text-gray-600" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <h3 className="font-medium text-gray-900">{integration.name}</h3>
                              <span className="text-xs text-gray-500">v{integration.version}</span>
                            </div>
                            <p className="text-sm text-gray-500 mt-0.5 line-clamp-2">
                              {integration.description}
                            </p>
                            <div className="flex items-center gap-2 mt-2">
                              <span className="text-xs px-2 py-0.5 bg-gray-100 rounded-full text-gray-600 capitalize">
                                {integration.category}
                              </span>
                              <span className="text-xs text-gray-400">
                                by {integration.author}
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

          {/* Right: Install form */}
          <div className="w-96 flex flex-col">
            {selectedIntegration ? (
              <>
                <div className="p-6 border-b border-gray-100">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-lg bg-primary-100 flex items-center justify-center">
                      <Icon className="w-6 h-6 text-primary-600" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-900">{selectedIntegration.name}</h3>
                      <p className="text-sm text-gray-500">{selectedIntegration.channelType}</p>
                    </div>
                  </div>
                  <p className="text-sm text-gray-600 mt-4">{selectedIntegration.description}</p>
                </div>

                <div className="flex-1 overflow-y-auto p-6">
                  <h4 className="text-sm font-medium text-gray-700 mb-4">Configuration</h4>

                  <div className="space-y-4">
                    {/* Slug input */}
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Integration Slug *
                      </label>
                      <input
                        type="text"
                        value={installSlug}
                        onChange={(e) => setInstallSlug(e.target.value)}
                        placeholder="my-slack-integration"
                        className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      />
                      <p className="text-xs text-gray-500 mt-1">Unique identifier for this integration</p>
                    </div>

                    {/* Dynamic config fields from schema */}
                    {Object.entries(selectedIntegration.configSchema.properties).map(([key, property]) => (
                      <div key={key}>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          {property.title}
                          {selectedIntegration.configSchema.required?.includes(key) && ' *'}
                        </label>
                        <input
                          type={property.secret ? 'password' : 'text'}
                          value={installConfig[key] || ''}
                          onChange={(e) => setInstallConfig({ ...installConfig, [key]: e.target.value })}
                          placeholder={property.description}
                          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        />
                        {property.description && (
                          <p className="text-xs text-gray-500 mt-1">{property.description}</p>
                        )}
                      </div>
                    ))}
                  </div>

                  {error && (
                    <div className="mt-4 p-3 bg-red-50 text-red-700 text-sm rounded-lg">
                      {error}
                    </div>
                  )}
                </div>

                <div className="p-4 border-t border-gray-200">
                  <button
                    onClick={handleInstall}
                    disabled={!installSlug || isInstalling}
                    className={clsx(
                      'w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                      !installSlug || isInstalling
                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                        : 'bg-primary-500 text-white hover:bg-primary-600'
                    )}
                  >
                    {isInstalling ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Installing...
                      </>
                    ) : (
                      <>
                        <CheckCircle className="w-4 h-4" />
                        Install Integration
                      </>
                    )}
                  </button>
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-gray-500">
                <p>Select an integration to install</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
