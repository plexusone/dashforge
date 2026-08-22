import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Clock,
  Copy,
  ExternalLink,
  LayoutDashboard,
  Loader2,
  Plus,
  Search,
  Trash2,
} from 'lucide-react'
import { listDashboards, deleteDashboard, duplicateDashboard } from '../../api/dashforge'
import { AnalyticsSourcePanel } from '../data-sources/AnalyticsSourcePanel'
import { AppShell } from '../navigation/AppShell'
import clsx from 'clsx'

/**
 * DashboardBrowsePage is the Dashboards section's landing page
 * (?mode=dashboards): the complete searchable dashboard list. The rail's
 * expandable Dashboards entry shows only the first few; this page is where
 * "view all" lands. Opening a dashboard navigates to the editor.
 */
export function DashboardBrowsePage() {
  const [search, setSearch] = useState('')
  const [showSourcePanel, setShowSourcePanel] = useState(false)

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['dashboards', 'browse', search],
    queryFn: () => listDashboards({ search: search || undefined, limit: 100 }),
  })

  const dashboards = data?.dashboards ?? []
  const total = data?.total ?? 0

  const handleDuplicate = async (id: string) => {
    try {
      await duplicateDashboard(id)
      await refetch()
    } catch (err) {
      console.error('Failed to duplicate dashboard:', err)
      alert(err instanceof Error ? err.message : 'Failed to duplicate dashboard')
    }
  }

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`Delete dashboard "${title}"? This cannot be undone.`)) return
    try {
      await deleteDashboard(id)
      await refetch()
    } catch (err) {
      console.error('Failed to delete dashboard:', err)
      alert(err instanceof Error ? err.message : 'Failed to delete dashboard')
    }
  }

  return (
    <AppShell active="dashboards" onOpenDataSources={() => setShowSourcePanel(true)}>
      <div className="h-full flex flex-col bg-gray-50">
        <header className="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-4 shadow-sm">
          <div>
            <h1 className="font-semibold text-gray-900 leading-tight">Dashboards</h1>
            <p className="text-xs text-gray-500">
              {isLoading ? 'Loading…' : `${total} dashboard${total === 1 ? '' : 's'}`}
            </p>
          </div>
          <a
            href="/builder/"
            className="inline-flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-primary-500 text-white hover:bg-primary-600"
          >
            <Plus className="w-4 h-4" />
            New dashboard
          </a>
        </header>

        <main className="flex-1 overflow-y-auto">
          <div className="max-w-3xl mx-auto p-6 space-y-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search dashboards..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm bg-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>

            {isLoading && (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="w-6 h-6 animate-spin text-primary-500" />
              </div>
            )}

            {error && !isLoading && (
              <div className="text-center py-12 text-sm">
                <p className="text-red-600 mb-2">
                  {error instanceof Error ? error.message : 'Failed to load dashboards'}
                </p>
                <button
                  onClick={() => refetch()}
                  className="text-primary-600 hover:text-primary-700"
                >
                  Try again
                </button>
              </div>
            )}

            {!isLoading && !error && dashboards.length === 0 && (
              <div className="text-center py-12 text-gray-500">
                <LayoutDashboard className="w-10 h-10 mx-auto mb-3 text-gray-300" />
                <p className="text-sm">
                  {search ? 'No dashboards match your search' : 'No dashboards yet'}
                </p>
                {!search && (
                  <a
                    href="/builder/"
                    className="mt-3 inline-block text-primary-600 hover:text-primary-700 text-sm"
                  >
                    Create your first dashboard
                  </a>
                )}
              </div>
            )}

            <div className="space-y-2">
              {dashboards.map((dashboard) => (
                <a
                  key={dashboard.id}
                  href={`/builder/?id=${dashboard.id}`}
                  className={clsx(
                    'group flex items-center justify-between px-4 py-3 bg-white border border-gray-200 rounded-lg',
                    'hover:border-gray-300 hover:shadow-sm transition-all',
                  )}
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="w-9 h-9 bg-gray-100 rounded flex items-center justify-center shrink-0">
                      <LayoutDashboard className="w-4 h-4 text-gray-400" />
                    </div>
                    <div className="min-w-0">
                      <div className="font-medium text-gray-900 truncate">{dashboard.title}</div>
                      <div className="flex items-center gap-3 text-xs text-gray-500 mt-0.5">
                        <span>{dashboard.widgets?.length || 0} widgets</span>
                        {dashboard.updatedAt && (
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {new Date(dashboard.updatedAt).toLocaleDateString()}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="hidden group-hover:flex items-center gap-1 shrink-0">
                    <button
                      onClick={(e) => {
                        e.preventDefault()
                        window.open(`/builder/?id=${dashboard.id}`, '_blank')
                      }}
                      className="p-1.5 rounded hover:bg-gray-100"
                      title="Open in new tab"
                    >
                      <ExternalLink className="w-4 h-4 text-gray-500" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.preventDefault()
                        handleDuplicate(dashboard.id)
                      }}
                      className="p-1.5 rounded hover:bg-gray-100"
                      title="Duplicate"
                    >
                      <Copy className="w-4 h-4 text-gray-500" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.preventDefault()
                        handleDelete(dashboard.id, dashboard.title)
                      }}
                      className="p-1.5 rounded hover:bg-red-100"
                      title="Delete"
                    >
                      <Trash2 className="w-4 h-4 text-gray-500 hover:text-red-500" />
                    </button>
                  </div>
                </a>
              ))}
            </div>
          </div>
        </main>

        {showSourcePanel && (
          <AnalyticsSourcePanel
            onClose={() => setShowSourcePanel(false)}
            onSourcesChanged={() => refetch()}
          />
        )}
      </div>
    </AppShell>
  )
}
