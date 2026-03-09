import { useState } from 'react'
import { Plus, Trash2, Database, Globe, Server, Code, X, ChevronDown, ChevronUp } from 'lucide-react'
import { useDashboardStore } from '../../stores/dashboard'
import type { DataSource, DataSourceType } from '../../types/dashboard'

interface DataSourcePanelProps {
  isOpen: boolean
  onClose: () => void
}

const DATA_SOURCE_ICONS: Record<DataSourceType, typeof Database> = {
  url: Globe,
  inline: Code,
  postgres: Database,
  mysql: Database,
  derived: Server,
  cube: Server
}

const DATA_SOURCE_LABELS: Record<DataSourceType, string> = {
  url: 'URL / API',
  inline: 'Inline Data',
  postgres: 'PostgreSQL',
  mysql: 'MySQL',
  derived: 'Derived',
  cube: 'Cube.js'
}

export function DataSourcePanel({ isOpen, onClose }: DataSourcePanelProps) {
  const { dashboard, addDataSource, updateDataSource, removeDataSource } = useDashboardStore()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <Database className="w-5 h-5 text-primary-500" />
            <h2 className="text-lg font-semibold text-gray-900">Data Sources</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Add Button */}
          {!showAddForm && (
            <button
              onClick={() => setShowAddForm(true)}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-primary-400 hover:text-primary-600 transition-colors mb-4"
            >
              <Plus className="w-4 h-4" />
              <span className="text-sm font-medium">Add Data Source</span>
            </button>
          )}

          {/* Add Form */}
          {showAddForm && (
            <DataSourceForm
              onSave={(ds) => {
                addDataSource(ds)
                setShowAddForm(false)
              }}
              onCancel={() => setShowAddForm(false)}
            />
          )}

          {/* Data Source List */}
          <div className="space-y-3">
            {dashboard.dataSources.length === 0 && !showAddForm && (
              <div className="text-center py-8 text-gray-500">
                <Database className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p className="text-sm">No data sources configured</p>
                <p className="text-xs mt-1">Add a data source to connect your widgets to data</p>
              </div>
            )}

            {dashboard.dataSources.map((ds) => (
              <DataSourceItem
                key={ds.id}
                dataSource={ds}
                isEditing={editingId === ds.id}
                onEdit={() => setEditingId(editingId === ds.id ? null : ds.id)}
                onUpdate={(updates) => {
                  updateDataSource(ds.id, updates)
                  setEditingId(null)
                }}
                onDelete={() => {
                  if (confirm(`Delete data source "${ds.id}"?`)) {
                    removeDataSource(ds.id)
                  }
                }}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

interface DataSourceItemProps {
  dataSource: DataSource
  isEditing: boolean
  onEdit: () => void
  onUpdate: (updates: Partial<DataSource>) => void
  onDelete: () => void
}

function DataSourceItem({ dataSource, isEditing, onEdit, onUpdate, onDelete }: DataSourceItemProps) {
  const Icon = DATA_SOURCE_ICONS[dataSource.type]

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 bg-gray-50 cursor-pointer hover:bg-gray-100 transition-colors"
        onClick={onEdit}
      >
        <div className="flex items-center gap-3">
          <div className="p-2 bg-white rounded-lg border border-gray-200">
            <Icon className="w-4 h-4 text-gray-600" />
          </div>
          <div>
            <p className="text-sm font-medium text-gray-900">{dataSource.id}</p>
            <p className="text-xs text-gray-500">{DATA_SOURCE_LABELS[dataSource.type]}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation()
              onDelete()
            }}
            className="p-1.5 hover:bg-red-100 rounded transition-colors"
            title="Delete"
          >
            <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-500" />
          </button>
          {isEditing ? (
            <ChevronUp className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          )}
        </div>
      </div>

      {/* Edit Form */}
      {isEditing && (
        <div className="p-4 border-t border-gray-200">
          <DataSourceEditForm
            dataSource={dataSource}
            onSave={onUpdate}
          />
        </div>
      )}
    </div>
  )
}

interface DataSourceFormProps {
  onSave: (dataSource: DataSource) => void
  onCancel: () => void
}

function DataSourceForm({ onSave, onCancel }: DataSourceFormProps) {
  const [id, setId] = useState('')
  const [type, setType] = useState<DataSourceType>('url')
  const [url, setUrl] = useState('')
  const [method, setMethod] = useState('GET')
  const [inlineData, setInlineData] = useState('')
  const [host, setHost] = useState('')
  const [port, setPort] = useState('')
  const [database, setDatabase] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [query, setQuery] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const dataSource: DataSource = {
      id: id || `ds-${Date.now()}`,
      type
    }

    if (type === 'url') {
      dataSource.url = url
      dataSource.method = method
    } else if (type === 'inline') {
      try {
        dataSource.data = JSON.parse(inlineData)
      } catch {
        dataSource.data = inlineData
      }
    } else if (type === 'postgres' || type === 'mysql') {
      dataSource.connection = {
        host,
        port: port ? parseInt(port, 10) : undefined,
        database,
        username,
        password
      }
      dataSource.query = query
    }

    onSave(dataSource)
  }

  return (
    <form onSubmit={handleSubmit} className="border border-primary-200 rounded-lg p-4 bg-primary-50/50 mb-4">
      <h3 className="text-sm font-medium text-gray-900 mb-3">New Data Source</h3>

      <div className="space-y-3">
        {/* ID */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">ID</label>
          <input
            type="text"
            value={id}
            onChange={(e) => setId(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            placeholder="my-data-source"
            required
          />
        </div>

        {/* Type */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Type</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as DataSourceType)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          >
            <option value="url">URL / API</option>
            <option value="inline">Inline Data</option>
            <option value="postgres">PostgreSQL</option>
            <option value="mysql">MySQL</option>
            <option value="cube">Cube.js</option>
          </select>
        </div>

        {/* Type-specific fields */}
        {type === 'url' && (
          <>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">URL</label>
              <input
                type="text"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                placeholder="https://api.example.com/data"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Method</label>
              <select
                value={method}
                onChange={(e) => setMethod(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
              </select>
            </div>
          </>
        )}

        {type === 'inline' && (
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Data (JSON)</label>
            <textarea
              value={inlineData}
              onChange={(e) => setInlineData(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono"
              rows={4}
              placeholder='[{"name": "Item 1", "value": 100}]'
            />
          </div>
        )}

        {(type === 'postgres' || type === 'mysql') && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Host</label>
                <input
                  type="text"
                  value={host}
                  onChange={(e) => setHost(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  placeholder="localhost"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Port</label>
                <input
                  type="text"
                  value={port}
                  onChange={(e) => setPort(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  placeholder={type === 'postgres' ? '5432' : '3306'}
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Database</label>
              <input
                type="text"
                value={database}
                onChange={(e) => setDatabase(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                placeholder="mydb"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Username</label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  placeholder="user"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Password</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  placeholder="••••••"
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Query</label>
              <textarea
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono"
                rows={3}
                placeholder="SELECT * FROM table"
              />
            </div>
          </>
        )}

        {type === 'cube' && (
          <div className="text-sm text-gray-500 bg-gray-100 rounded-lg p-3">
            Cube.js data sources are configured via the Query Builder panel when editing a chart widget.
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-2 mt-4">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-3 py-1.5 text-sm bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          Add Data Source
        </button>
      </div>
    </form>
  )
}

interface DataSourceEditFormProps {
  dataSource: DataSource
  onSave: (updates: Partial<DataSource>) => void
}

function DataSourceEditForm({ dataSource, onSave }: DataSourceEditFormProps) {
  const [url, setUrl] = useState(dataSource.url || '')
  const [method, setMethod] = useState(dataSource.method || 'GET')
  const [inlineData, setInlineData] = useState(
    dataSource.data ? JSON.stringify(dataSource.data, null, 2) : ''
  )
  const [host, setHost] = useState(dataSource.connection?.host || '')
  const [port, setPort] = useState(dataSource.connection?.port?.toString() || '')
  const [database, setDatabase] = useState(dataSource.connection?.database || '')
  const [username, setUsername] = useState(dataSource.connection?.username || '')
  const [password, setPassword] = useState(dataSource.connection?.password || '')
  const [query, setQuery] = useState(dataSource.query || '')

  const handleSave = () => {
    const updates: Partial<DataSource> = {}

    if (dataSource.type === 'url') {
      updates.url = url
      updates.method = method
    } else if (dataSource.type === 'inline') {
      try {
        updates.data = JSON.parse(inlineData)
      } catch {
        updates.data = inlineData
      }
    } else if (dataSource.type === 'postgres' || dataSource.type === 'mysql') {
      updates.connection = {
        host,
        port: port ? parseInt(port, 10) : undefined,
        database,
        username,
        password
      }
      updates.query = query
    }

    onSave(updates)
  }

  return (
    <div className="space-y-3">
      {dataSource.type === 'url' && (
        <>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">URL</label>
            <input
              type="text"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              placeholder="https://api.example.com/data"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Method</label>
            <select
              value={method}
              onChange={(e) => setMethod(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
            </select>
          </div>
        </>
      )}

      {dataSource.type === 'inline' && (
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Data (JSON)</label>
          <textarea
            value={inlineData}
            onChange={(e) => setInlineData(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono"
            rows={4}
          />
        </div>
      )}

      {(dataSource.type === 'postgres' || dataSource.type === 'mysql') && (
        <>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Host</label>
              <input
                type="text"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Port</label>
              <input
                type="text"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Database</label>
            <input
              type="text"
              value={database}
              onChange={(e) => setDatabase(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Query</label>
            <textarea
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono"
              rows={3}
            />
          </div>
        </>
      )}

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          className="px-3 py-1.5 text-sm bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          Save Changes
        </button>
      </div>
    </div>
  )
}
