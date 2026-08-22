import { ReactNode, useState, useCallback, useEffect } from 'react'
import {
  Save,
  Undo2,
  Redo2,
  Play,
  Eye,
  Settings,
  Download,
  Upload,
  HelpCircle,
  Loader2,
  Check,
  AlertCircle,
  Database,
  Variable,
} from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useDashboardStore } from '../stores/dashboard'
import { createDashboard, updateDashboard } from '../api/dashforge'
import { DataSourcePanel } from './data-source'
import { AnalyticsSourcePanel } from './data-sources/AnalyticsSourcePanel'
import { AppShell } from './navigation/AppShell'
import { VariablePanel } from './variables'
import { MarketplacePanel, IntegrationSettingsPanel } from './integrations'
import { AlertPanel } from './alerts'
import clsx from 'clsx'

interface LayoutProps {
  children: ReactNode
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error'

export function Layout({ children }: LayoutProps) {
  const [showMenu, setShowMenu] = useState(false)
  const [showDataSources, setShowDataSources] = useState(false)
  const [showAnalyticsSources, setShowAnalyticsSources] = useState(false)
  const [showVariables, setShowVariables] = useState(false)
  const [showMarketplace, setShowMarketplace] = useState(false)
  const [showIntegrations, setShowIntegrations] = useState(false)
  const [showAlerts, setShowAlerts] = useState(false)
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle')
  const [saveError, setSaveError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  useEffect(() => {
    const warnUnsaved = (e: BeforeUnloadEvent) => {
      if (useDashboardStore.getState().isDirty) {
        e.preventDefault()
      }
    }
    window.addEventListener('beforeunload', warnUnsaved)
    return () => window.removeEventListener('beforeunload', warnUnsaved)
  }, [])

  const {
    dashboard,
    isDirty,
    isEditing,
    setEditMode,
    undo,
    redo,
    canUndo,
    canRedo,
    exportDashboard,
    setDashboard,
    markClean,
  } = useDashboardStore()

  const handleSave = useCallback(async () => {
    if (saveStatus === 'saving') return

    setSaveStatus('saving')
    setSaveError(null)

    try {
      const dashboardData = exportDashboard()
      let savedDashboard

      if (dashboard.id && !dashboard.id.startsWith('new-')) {
        // Update existing dashboard
        savedDashboard = await updateDashboard(dashboard.id, dashboardData)
      } else {
        // Create new dashboard
        savedDashboard = await createDashboard(dashboardData)
        // Update the local dashboard with the new ID
        setDashboard({ ...dashboard, id: savedDashboard.id })
      }

      markClean()
      setSaveStatus('saved')
      void queryClient.invalidateQueries({ queryKey: ['dashboards'] })

      // Reset status after 2 seconds
      setTimeout(() => setSaveStatus('idle'), 2000)
    } catch (err) {
      console.error('Failed to save dashboard:', err)
      setSaveError(err instanceof Error ? err.message : 'Failed to save')
      setSaveStatus('error')

      // Reset error status after 3 seconds
      setTimeout(() => {
        setSaveStatus('idle')
        setSaveError(null)
      }, 3000)
    }
  }, [dashboard, exportDashboard, setDashboard, markClean, saveStatus, queryClient])

  const handleExport = () => {
    const data = exportDashboard()
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${dashboard.title.toLowerCase().replace(/\s+/g, '-')}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleImport = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0]
      if (file) {
        const text = await file.text()
        try {
          const data = JSON.parse(text)
          useDashboardStore.getState().setDashboard(data)
        } catch (err) {
          console.error('Failed to parse dashboard JSON:', err)
          alert('Invalid dashboard file')
        }
      }
    }
    input.click()
  }

  return (
    <AppShell
      active="dashboards"
      onOpenDataSources={() => setShowAnalyticsSources(true)}
      onOpenIntegrations={() => setShowIntegrations(true)}
      onOpenAlerts={() => setShowAlerts(true)}
    >
      <div className="h-full flex bg-gray-50">
        <div className="flex-1 min-w-0 flex flex-col">
          {/* Header */}
          <header className="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-4 shadow-sm">
            <div className="flex items-center gap-3">
              <div>
                <h1 className="font-semibold text-gray-900 leading-tight">
                  {dashboard.title}
                  {isDirty && <span className="text-primary-500 ml-1">*</span>}
                </h1>
                <p className="text-xs text-gray-500">Dashboard Builder</p>
              </div>
            </div>

            {/* Toolbar */}
            <div className="flex items-center gap-2">
              {/* Undo/Redo */}
              <div className="flex items-center border-r border-gray-200 pr-2 mr-2">
                <button
                  onClick={undo}
                  disabled={!canUndo()}
                  className={clsx(
                    'p-2 rounded-lg transition-colors',
                    canUndo() ? 'hover:bg-gray-100' : 'opacity-40 cursor-not-allowed',
                  )}
                  title="Undo (Ctrl+Z)"
                >
                  <Undo2 className="w-4 h-4 text-gray-600" />
                </button>
                <button
                  onClick={redo}
                  disabled={!canRedo()}
                  className={clsx(
                    'p-2 rounded-lg transition-colors',
                    canRedo() ? 'hover:bg-gray-100' : 'opacity-40 cursor-not-allowed',
                  )}
                  title="Redo (Ctrl+Shift+Z)"
                >
                  <Redo2 className="w-4 h-4 text-gray-600" />
                </button>
              </div>

              {/* View Mode Toggle */}
              <button
                onClick={() => setEditMode(!isEditing)}
                className={clsx(
                  'flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors',
                  isEditing ? 'bg-gray-100 text-gray-700' : 'bg-primary-50 text-primary-700',
                )}
              >
                {isEditing ? (
                  <>
                    <Eye className="w-4 h-4" />
                    <span className="text-sm">Preview</span>
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4" />
                    <span className="text-sm">Edit</span>
                  </>
                )}
              </button>

              {/* Save */}
              <button
                onClick={handleSave}
                disabled={saveStatus === 'saving'}
                className={clsx(
                  'flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors',
                  saveStatus === 'error'
                    ? 'bg-red-500 text-white hover:bg-red-600'
                    : saveStatus === 'saved'
                      ? 'bg-green-500 text-white'
                      : 'bg-primary-500 text-white hover:bg-primary-600',
                  saveStatus === 'saving' && 'opacity-75 cursor-wait',
                )}
                title={saveError || undefined}
              >
                {saveStatus === 'saving' ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : saveStatus === 'saved' ? (
                  <Check className="w-4 h-4" />
                ) : saveStatus === 'error' ? (
                  <AlertCircle className="w-4 h-4" />
                ) : (
                  <Save className="w-4 h-4" />
                )}
                <span className="text-sm">
                  {saveStatus === 'saving'
                    ? 'Saving...'
                    : saveStatus === 'saved'
                      ? 'Saved!'
                      : saveStatus === 'error'
                        ? 'Error'
                        : 'Save'}
                </span>
              </button>

              {/* More Actions */}
              <div className="relative">
                <button
                  onClick={() => setShowMenu(!showMenu)}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  <Settings className="w-4 h-4 text-gray-600" />
                </button>

                {showMenu && (
                  <div className="absolute right-0 top-full mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
                    <div className="px-4 pt-1 pb-1 text-xs font-medium text-gray-400 uppercase tracking-wide">
                      Dashboard settings
                    </div>
                    <button
                      onClick={() => {
                        setShowDataSources(true)
                        setShowMenu(false)
                      }}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      <Database className="w-4 h-4" />
                      Dashboard data
                    </button>
                    <button
                      onClick={() => {
                        setShowVariables(true)
                        setShowMenu(false)
                      }}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      <Variable className="w-4 h-4" />
                      Variables
                    </button>
                    <hr className="my-1 border-gray-200" />
                    <button
                      onClick={handleExport}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      <Download className="w-4 h-4" />
                      Export JSON
                    </button>
                    <button
                      onClick={handleImport}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      <Upload className="w-4 h-4" />
                      Import JSON
                    </button>
                    <hr className="my-1 border-gray-200" />
                    <button className="w-full flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                      <HelpCircle className="w-4 h-4" />
                      Help
                    </button>
                  </div>
                )}
              </div>
            </div>
          </header>

          {/* Main Content */}
          <main className="flex-1 overflow-hidden">{children}</main>

          {/* Data Source Panel */}
          <DataSourcePanel isOpen={showDataSources} onClose={() => setShowDataSources(false)} />

          {/* Variable Panel */}
          <VariablePanel isOpen={showVariables} onClose={() => setShowVariables(false)} />

          {/* Analytics Source Panel */}
          {showAnalyticsSources && (
            <AnalyticsSourcePanel
              onClose={() => setShowAnalyticsSources(false)}
              onSourcesChanged={() => {}}
            />
          )}

          {/* Marketplace Panel */}
          <MarketplacePanel isOpen={showMarketplace} onClose={() => setShowMarketplace(false)} />

          {/* Integration Settings Panel */}
          <IntegrationSettingsPanel
            isOpen={showIntegrations}
            onClose={() => setShowIntegrations(false)}
          />

          {/* Alert Panel */}
          <AlertPanel isOpen={showAlerts} onClose={() => setShowAlerts(false)} />
        </div>
      </div>
    </AppShell>
  )
}
