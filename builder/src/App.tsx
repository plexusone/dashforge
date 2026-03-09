import { useState, useEffect } from 'react'
import { Layout } from './components/Layout'
import { Canvas } from './components/canvas/Canvas'
import { WidgetPalette } from './components/palette/WidgetPalette'
import { PropertiesPanel } from './components/properties/PropertiesPanel'
import { AIChat, AICommandBar, useAICommandBar } from './components/ai'
import { useDashboardStore } from './stores/dashboard'
import { getDashboard } from './api/dashforge'
import { Loader2 } from 'lucide-react'

function App() {
  const [selectedWidgetId, setSelectedWidgetId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)
  const { setDashboard } = useDashboardStore()
  const aiCommandBar = useAICommandBar()

  // Load dashboard from URL params or create new
  useEffect(() => {
    const loadDashboard = async () => {
      const params = new URLSearchParams(window.location.search)
      const dashboardId = params.get('id')

      if (dashboardId) {
        try {
          const dashboard = await getDashboard(dashboardId)
          setDashboard(dashboard)
        } catch (err) {
          console.error('Failed to load dashboard:', err)
          setLoadError(err instanceof Error ? err.message : 'Failed to load dashboard')
        }
      }
      setIsLoading(false)
    }

    loadDashboard()
  }, [setDashboard])

  // Show loading state
  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin text-primary-500 mx-auto mb-4" />
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  // Show error state with option to create new
  if (loadError) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center max-w-md">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-red-500 text-xl">!</span>
          </div>
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Failed to load dashboard</h2>
          <p className="text-gray-600 mb-4">{loadError}</p>
          <button
            onClick={() => {
              setLoadError(null)
              window.history.replaceState({}, '', window.location.pathname)
            }}
            className="px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600"
          >
            Create New Dashboard
          </button>
        </div>
      </div>
    )
  }

  return (
    <Layout>
      <div className="flex h-full">
        {/* Left Sidebar - Widget Palette */}
        <WidgetPalette />

        {/* Main Canvas */}
        <div className="flex-1 overflow-auto bg-gray-100 p-4">
          <Canvas
            selectedWidgetId={selectedWidgetId}
            onSelectWidget={setSelectedWidgetId}
          />
        </div>

        {/* Right Sidebar - Properties Panel */}
        <PropertiesPanel
          selectedWidgetId={selectedWidgetId}
          onClose={() => setSelectedWidgetId(null)}
        />
      </div>

      {/* AI Components */}
      <AIChat />
      <AICommandBar
        isOpen={aiCommandBar.isOpen}
        onClose={aiCommandBar.close}
      />
    </Layout>
  )
}

export default App
