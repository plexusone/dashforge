import { useMemo } from 'react'
import { Sidebar, SidebarSection } from '../Sidebar'
import { useDashboardStore } from '../../stores/dashboard'
import { ChartBuilder } from '../chart-builder/ChartBuilder'
import type { ChartConfig, MetricConfig, TableConfig, TextConfig, ImageConfig } from '../../types/dashboard'

interface PropertiesPanelProps {
  selectedWidgetId: string | null
  onClose: () => void
}

export function PropertiesPanel({ selectedWidgetId, onClose }: PropertiesPanelProps) {
  const { dashboard, updateWidget } = useDashboardStore()

  const selectedWidget = useMemo(() =>
    dashboard.widgets.find(w => w.id === selectedWidgetId),
    [dashboard.widgets, selectedWidgetId]
  )

  if (!selectedWidgetId || !selectedWidget) {
    return (
      <Sidebar title="Properties" position="right" width="w-80">
        <div className="p-4 text-center text-gray-500">
          <p className="text-sm">Select a widget to edit its properties</p>
        </div>
      </Sidebar>
    )
  }

  const handleTitleChange = (title: string) => {
    updateWidget(selectedWidgetId, { title })
  }

  const handleConfigChange = (config: unknown) => {
    updateWidget(selectedWidgetId, { config })
  }

  const handleDataSourceChange = (datasourceId: string) => {
    updateWidget(selectedWidgetId, { datasourceId })
  }

  return (
    <Sidebar title="Properties" position="right" width="w-80" onClose={onClose}>
      {/* General Settings */}
      <SidebarSection title="General">
        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">
              Title
            </label>
            <input
              type="text"
              value={selectedWidget.title || ''}
              onChange={(e) => handleTitleChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="Widget title"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">
              Description
            </label>
            <textarea
              value={selectedWidget.description || ''}
              onChange={(e) => updateWidget(selectedWidgetId, { description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              placeholder="Optional description"
              rows={2}
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">
              Data Source
            </label>
            <select
              value={selectedWidget.datasourceId || ''}
              onChange={(e) => handleDataSourceChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            >
              <option value="">None</option>
              {dashboard.dataSources.map(ds => (
                <option key={ds.id} value={ds.id}>{ds.id}</option>
              ))}
            </select>
          </div>
        </div>
      </SidebarSection>

      {/* Widget-specific config */}
      {selectedWidget.type === 'chart' && (
        <SidebarSection title="Chart Configuration">
          <ChartBuilder
            config={selectedWidget.config as ChartConfig}
            onChange={handleConfigChange}
          />
        </SidebarSection>
      )}

      {selectedWidget.type === 'metric' && (
        <MetricConfigEditor
          config={selectedWidget.config as MetricConfig}
          onChange={handleConfigChange}
        />
      )}

      {selectedWidget.type === 'table' && (
        <TableConfigEditor
          config={selectedWidget.config as TableConfig}
          onChange={handleConfigChange}
        />
      )}

      {selectedWidget.type === 'text' && (
        <TextConfigEditor
          config={selectedWidget.config as TextConfig}
          onChange={handleConfigChange}
        />
      )}

      {selectedWidget.type === 'image' && (
        <ImageConfigEditor
          config={selectedWidget.config as ImageConfig}
          onChange={handleConfigChange}
        />
      )}

      {/* Position info (read-only) */}
      <SidebarSection title="Position">
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div className="bg-gray-50 p-2 rounded">
            <span className="text-gray-500">X:</span> {selectedWidget.position.x}
          </div>
          <div className="bg-gray-50 p-2 rounded">
            <span className="text-gray-500">Y:</span> {selectedWidget.position.y}
          </div>
          <div className="bg-gray-50 p-2 rounded">
            <span className="text-gray-500">Width:</span> {selectedWidget.position.w}
          </div>
          <div className="bg-gray-50 p-2 rounded">
            <span className="text-gray-500">Height:</span> {selectedWidget.position.h}
          </div>
        </div>
      </SidebarSection>
    </Sidebar>
  )
}

// Metric Config Editor
interface MetricConfigEditorProps {
  config: MetricConfig
  onChange: (config: MetricConfig) => void
}

function MetricConfigEditor({ config, onChange }: MetricConfigEditorProps) {
  return (
    <SidebarSection title="Metric Configuration">
      <div className="space-y-3">
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Value Field
          </label>
          <input
            type="text"
            value={config.valueField || ''}
            onChange={(e) => onChange({ ...config, valueField: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            placeholder="e.g., total_revenue"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Format
          </label>
          <select
            value={config.format || 'number'}
            onChange={(e) => onChange({ ...config, format: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="number">Number</option>
            <option value="currency">Currency</option>
            <option value="percent">Percent</option>
            <option value="compact">Compact</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="show-comparison"
            checked={config.comparison?.enabled || false}
            onChange={(e) => onChange({
              ...config,
              comparison: { ...config.comparison, enabled: e.target.checked }
            })}
            className="rounded border-gray-300"
          />
          <label htmlFor="show-comparison" className="text-sm text-gray-600">
            Show comparison
          </label>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="show-sparkline"
            checked={config.sparkline?.enabled || false}
            onChange={(e) => onChange({
              ...config,
              sparkline: { ...config.sparkline, enabled: e.target.checked }
            })}
            className="rounded border-gray-300"
          />
          <label htmlFor="show-sparkline" className="text-sm text-gray-600">
            Show sparkline
          </label>
        </div>
      </div>
    </SidebarSection>
  )
}

// Table Config Editor
interface TableConfigEditorProps {
  config: TableConfig
  onChange: (config: TableConfig) => void
}

function TableConfigEditor({ config, onChange }: TableConfigEditorProps) {
  return (
    <SidebarSection title="Table Configuration">
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="pagination"
            checked={config.pagination?.enabled || false}
            onChange={(e) => onChange({
              ...config,
              pagination: { ...config.pagination, enabled: e.target.checked }
            })}
            className="rounded border-gray-300"
          />
          <label htmlFor="pagination" className="text-sm text-gray-600">
            Enable pagination
          </label>
        </div>

        {config.pagination?.enabled && (
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">
              Page Size
            </label>
            <input
              type="number"
              value={config.pagination?.pageSize || 10}
              onChange={(e) => onChange({
                ...config,
                pagination: { ...config.pagination, enabled: true, pageSize: parseInt(e.target.value) }
              })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              min={1}
              max={100}
            />
          </div>
        )}

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="sortable"
            checked={config.sortable || false}
            onChange={(e) => onChange({ ...config, sortable: e.target.checked })}
            className="rounded border-gray-300"
          />
          <label htmlFor="sortable" className="text-sm text-gray-600">
            Sortable columns
          </label>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="filterable"
            checked={config.filterable || false}
            onChange={(e) => onChange({ ...config, filterable: e.target.checked })}
            className="rounded border-gray-300"
          />
          <label htmlFor="filterable" className="text-sm text-gray-600">
            Filterable columns
          </label>
        </div>
      </div>
    </SidebarSection>
  )
}

// Text Config Editor
interface TextConfigEditorProps {
  config: TextConfig
  onChange: (config: TextConfig) => void
}

function TextConfigEditor({ config, onChange }: TextConfigEditorProps) {
  return (
    <SidebarSection title="Text Configuration">
      <div className="space-y-3">
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Format
          </label>
          <select
            value={config.format || 'plain'}
            onChange={(e) => onChange({ ...config, format: e.target.value as TextConfig['format'] })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="plain">Plain Text</option>
            <option value="markdown">Markdown</option>
            <option value="html">HTML</option>
          </select>
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Content
          </label>
          <textarea
            value={config.content || ''}
            onChange={(e) => onChange({ ...config, content: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono"
            rows={6}
            placeholder="Enter text content..."
          />
        </div>
      </div>
    </SidebarSection>
  )
}

// Image Config Editor
interface ImageConfigEditorProps {
  config: ImageConfig
  onChange: (config: ImageConfig) => void
}

function ImageConfigEditor({ config, onChange }: ImageConfigEditorProps) {
  return (
    <SidebarSection title="Image Configuration">
      <div className="space-y-3">
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Image URL
          </label>
          <input
            type="text"
            value={config.src || ''}
            onChange={(e) => onChange({ ...config, src: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            placeholder="https://example.com/image.png"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Alt Text
          </label>
          <input
            type="text"
            value={config.alt || ''}
            onChange={(e) => onChange({ ...config, alt: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            placeholder="Image description for accessibility"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Fit Mode
          </label>
          <select
            value={config.fit || 'contain'}
            onChange={(e) => onChange({ ...config, fit: e.target.value as ImageConfig['fit'] })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="contain">Contain (fit within bounds)</option>
            <option value="cover">Cover (fill bounds, may crop)</option>
            <option value="fill">Fill (stretch to fit)</option>
            <option value="none">None (original size)</option>
          </select>
        </div>
      </div>
    </SidebarSection>
  )
}
