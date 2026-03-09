import { useState, useEffect, useCallback } from 'react'
import {
  Search,
  Plus,
  LayoutDashboard,
  X,
  Trash2,
  Copy,
  ExternalLink,
  Clock,
  Loader2,
  Grid3X3,
  List
} from 'lucide-react'
import { listDashboards, deleteDashboard, duplicateDashboard, ListDashboardsResponse } from '../../api/dashforge'
import { useDashboardStore } from '../../stores/dashboard'
import type { Dashboard } from '../../types/dashboard'
import clsx from 'clsx'

interface DashboardGalleryProps {
  isOpen: boolean
  onClose: () => void
  onSelectDashboard: (dashboard: Dashboard) => void
  onCreateNew: () => void
}

export function DashboardGallery({ isOpen, onClose, onSelectDashboard, onCreateNew }: DashboardGalleryProps) {
  const [dashboards, setDashboards] = useState<Dashboard[]>([])
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [total, setTotal] = useState(0)

  const { dashboard: currentDashboard, isDirty } = useDashboardStore()

  const loadDashboards = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result: ListDashboardsResponse = await listDashboards({
        search: search || undefined,
        limit: 50
      })
      setDashboards(result.dashboards)
      setTotal(result.total)
    } catch (err) {
      console.error('Failed to load dashboards:', err)
      setError(err instanceof Error ? err.message : 'Failed to load dashboards')
    } finally {
      setIsLoading(false)
    }
  }, [search])

  useEffect(() => {
    if (isOpen) {
      loadDashboards()
    }
  }, [isOpen, loadDashboards])

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`Delete dashboard "${title}"? This cannot be undone.`)) return

    try {
      await deleteDashboard(id)
      setDashboards(dashboards.filter(d => d.id !== id))
      setTotal(total - 1)
    } catch (err) {
      console.error('Failed to delete dashboard:', err)
      alert(err instanceof Error ? err.message : 'Failed to delete dashboard')
    }
  }

  const handleDuplicate = async (id: string) => {
    try {
      const newDashboard = await duplicateDashboard(id)
      setDashboards([newDashboard, ...dashboards])
      setTotal(total + 1)
    } catch (err) {
      console.error('Failed to duplicate dashboard:', err)
      alert(err instanceof Error ? err.message : 'Failed to duplicate dashboard')
    }
  }

  const handleSelect = (dashboard: Dashboard) => {
    if (isDirty) {
      if (!confirm('You have unsaved changes. Discard them and open another dashboard?')) {
        return
      }
    }
    onSelectDashboard(dashboard)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-4xl max-h-[85vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <LayoutDashboard className="w-5 h-5 text-primary-500" />
            <h2 className="text-lg font-semibold text-gray-900">Dashboards</h2>
            {!isLoading && (
              <span className="text-sm text-gray-500">({total} total)</span>
            )}
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Toolbar */}
        <div className="flex items-center justify-between px-6 py-3 border-b border-gray-100">
          {/* Search */}
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search dashboards..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            />
          </div>

          <div className="flex items-center gap-2">
            {/* View Mode Toggle */}
            <div className="flex items-center border border-gray-300 rounded-lg overflow-hidden">
              <button
                onClick={() => setViewMode('grid')}
                className={clsx(
                  'p-2 transition-colors',
                  viewMode === 'grid' ? 'bg-gray-100 text-gray-900' : 'text-gray-500 hover:bg-gray-50'
                )}
                title="Grid view"
              >
                <Grid3X3 className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={clsx(
                  'p-2 transition-colors',
                  viewMode === 'list' ? 'bg-gray-100 text-gray-900' : 'text-gray-500 hover:bg-gray-50'
                )}
                title="List view"
              >
                <List className="w-4 h-4" />
              </button>
            </div>

            {/* Create New */}
            <button
              onClick={() => {
                onCreateNew()
                onClose()
              }}
              className="flex items-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span className="text-sm font-medium">New Dashboard</span>
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {isLoading && (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary-500" />
            </div>
          )}

          {error && !isLoading && (
            <div className="text-center py-12">
              <div className="text-red-500 mb-4">{error}</div>
              <button
                onClick={loadDashboards}
                className="text-primary-600 hover:text-primary-700 text-sm"
              >
                Try again
              </button>
            </div>
          )}

          {!isLoading && !error && dashboards.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              <LayoutDashboard className="w-12 h-12 mx-auto mb-3 text-gray-300" />
              <p className="text-sm">
                {search ? 'No dashboards match your search' : 'No dashboards yet'}
              </p>
              {!search && (
                <button
                  onClick={() => {
                    onCreateNew()
                    onClose()
                  }}
                  className="mt-4 text-primary-600 hover:text-primary-700 text-sm"
                >
                  Create your first dashboard
                </button>
              )}
            </div>
          )}

          {!isLoading && !error && dashboards.length > 0 && (
            viewMode === 'grid' ? (
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {dashboards.map((dashboard) => (
                  <DashboardCard
                    key={dashboard.id}
                    dashboard={dashboard}
                    isCurrent={dashboard.id === currentDashboard.id}
                    onSelect={() => handleSelect(dashboard)}
                    onDuplicate={() => handleDuplicate(dashboard.id)}
                    onDelete={() => handleDelete(dashboard.id, dashboard.title)}
                  />
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                {dashboards.map((dashboard) => (
                  <DashboardListItem
                    key={dashboard.id}
                    dashboard={dashboard}
                    isCurrent={dashboard.id === currentDashboard.id}
                    onSelect={() => handleSelect(dashboard)}
                    onDuplicate={() => handleDuplicate(dashboard.id)}
                    onDelete={() => handleDelete(dashboard.id, dashboard.title)}
                  />
                ))}
              </div>
            )
          )}
        </div>
      </div>
    </div>
  )
}

interface DashboardCardProps {
  dashboard: Dashboard
  isCurrent: boolean
  onSelect: () => void
  onDuplicate: () => void
  onDelete: () => void
}

function DashboardCard({ dashboard, isCurrent, onSelect, onDuplicate, onDelete }: DashboardCardProps) {
  return (
    <div
      className={clsx(
        'group relative border rounded-lg overflow-hidden cursor-pointer transition-all hover:shadow-md',
        isCurrent ? 'border-primary-500 ring-2 ring-primary-200' : 'border-gray-200 hover:border-gray-300'
      )}
      onClick={onSelect}
    >
      {/* Preview Area */}
      <div className="h-32 bg-gray-50 flex items-center justify-center relative">
        <div className="grid grid-cols-3 gap-1 p-4 w-full h-full opacity-30">
          {/* Mock widget previews */}
          {dashboard.widgets?.slice(0, 6).map((_, i) => (
            <div key={i} className="bg-gray-300 rounded" />
          ))}
          {(dashboard.widgets?.length || 0) < 6 &&
            Array(6 - (dashboard.widgets?.length || 0)).fill(0).map((_, i) => (
              <div key={`empty-${i}`} className="bg-gray-200 rounded border border-dashed border-gray-300" />
            ))
          }
        </div>

        {/* Widget count badge */}
        <div className="absolute bottom-2 right-2 bg-white px-2 py-0.5 rounded text-xs text-gray-500 shadow-sm">
          {dashboard.widgets?.length || 0} widgets
        </div>

        {isCurrent && (
          <div className="absolute top-2 left-2 bg-primary-500 text-white px-2 py-0.5 rounded text-xs">
            Current
          </div>
        )}
      </div>

      {/* Info */}
      <div className="p-3">
        <h3 className="font-medium text-gray-900 truncate">{dashboard.title}</h3>
        {dashboard.description && (
          <p className="text-xs text-gray-500 truncate mt-1">{dashboard.description}</p>
        )}
      </div>

      {/* Actions (visible on hover) */}
      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
        <button
          onClick={(e) => {
            e.stopPropagation()
            onDuplicate()
          }}
          className="p-1.5 bg-white rounded shadow hover:bg-gray-50"
          title="Duplicate"
        >
          <Copy className="w-3 h-3 text-gray-600" />
        </button>
        <button
          onClick={(e) => {
            e.stopPropagation()
            onDelete()
          }}
          className="p-1.5 bg-white rounded shadow hover:bg-red-50"
          title="Delete"
        >
          <Trash2 className="w-3 h-3 text-gray-600 hover:text-red-500" />
        </button>
      </div>
    </div>
  )
}

interface DashboardListItemProps {
  dashboard: Dashboard
  isCurrent: boolean
  onSelect: () => void
  onDuplicate: () => void
  onDelete: () => void
}

function DashboardListItem({ dashboard, isCurrent, onSelect, onDuplicate, onDelete }: DashboardListItemProps) {
  return (
    <div
      className={clsx(
        'flex items-center justify-between px-4 py-3 border rounded-lg cursor-pointer transition-all hover:shadow-sm',
        isCurrent ? 'border-primary-500 bg-primary-50' : 'border-gray-200 hover:border-gray-300'
      )}
      onClick={onSelect}
    >
      <div className="flex items-center gap-4 min-w-0">
        <div className="w-10 h-10 bg-gray-100 rounded flex items-center justify-center shrink-0">
          <LayoutDashboard className="w-5 h-5 text-gray-400" />
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-gray-900 truncate">{dashboard.title}</h3>
            {isCurrent && (
              <span className="px-2 py-0.5 bg-primary-500 text-white rounded text-xs shrink-0">
                Current
              </span>
            )}
          </div>
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

      <div className="flex items-center gap-1 shrink-0">
        <button
          onClick={(e) => {
            e.stopPropagation()
            window.open(`?id=${dashboard.id}`, '_blank')
          }}
          className="p-2 hover:bg-gray-100 rounded transition-colors"
          title="Open in new tab"
        >
          <ExternalLink className="w-4 h-4 text-gray-400" />
        </button>
        <button
          onClick={(e) => {
            e.stopPropagation()
            onDuplicate()
          }}
          className="p-2 hover:bg-gray-100 rounded transition-colors"
          title="Duplicate"
        >
          <Copy className="w-4 h-4 text-gray-400" />
        </button>
        <button
          onClick={(e) => {
            e.stopPropagation()
            onDelete()
          }}
          className="p-2 hover:bg-red-100 rounded transition-colors"
          title="Delete"
        >
          <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-500" />
        </button>
      </div>
    </div>
  )
}
