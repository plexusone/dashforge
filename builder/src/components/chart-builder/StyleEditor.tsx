import type { ChartConfig, GeometryType } from '../../types/dashboard'

interface StyleEditorProps {
  geometry: GeometryType
  style: ChartConfig['style']
  onChange: (style: ChartConfig['style']) => void
}

const colorPresets = [
  ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de'],
  ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'],
  ['#0ea5e9', '#22c55e', '#eab308', '#f43f5e', '#a855f7'],
  ['#6366f1', '#14b8a6', '#f97316', '#ec4899', '#84cc16']
]

export function StyleEditor({ geometry, style, onChange }: StyleEditorProps) {
  const handleChange = <K extends keyof NonNullable<ChartConfig['style']>>(
    key: K,
    value: NonNullable<ChartConfig['style']>[K]
  ) => {
    onChange({
      ...style,
      [key]: value
    })
  }

  const isPieType = geometry === 'pie' || geometry === 'funnel'
  const isLineType = geometry === 'line' || geometry === 'area'
  const isBarType = geometry === 'bar'
  const isAreaType = geometry === 'area'

  return (
    <div className="space-y-4">
      <label className="block text-xs font-medium text-gray-600">
        Style Options
      </label>

      {/* Color Palette */}
      <div>
        <label className="block text-xs text-gray-500 mb-2">Color Palette</label>
        <div className="space-y-2">
          {colorPresets.map((palette, idx) => (
            <button
              key={idx}
              onClick={() => handleChange('colors', palette)}
              className={`flex gap-1 p-1 rounded border ${
                JSON.stringify(style?.colors) === JSON.stringify(palette)
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              {palette.map((color) => (
                <div
                  key={color}
                  className="w-5 h-5 rounded"
                  style={{ backgroundColor: color }}
                />
              ))}
            </button>
          ))}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center justify-between">
        <label className="text-sm text-gray-600">Show Legend</label>
        <input
          type="checkbox"
          checked={style?.showLegend !== false}
          onChange={(e) => handleChange('showLegend', e.target.checked)}
          className="rounded border-gray-300"
        />
      </div>

      {style?.showLegend !== false && (
        <div>
          <label className="block text-xs text-gray-500 mb-1">Legend Position</label>
          <select
            value={style?.legendPosition || 'top'}
            onChange={(e) => handleChange('legendPosition', e.target.value as 'top' | 'bottom' | 'left' | 'right')}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="top">Top</option>
            <option value="bottom">Bottom</option>
            <option value="left">Left</option>
            <option value="right">Right</option>
          </select>
        </div>
      )}

      {/* Labels (for pie charts) */}
      {isPieType && (
        <div className="flex items-center justify-between">
          <label className="text-sm text-gray-600">Show Labels</label>
          <input
            type="checkbox"
            checked={style?.showLabels !== false}
            onChange={(e) => handleChange('showLabels', e.target.checked)}
            className="rounded border-gray-300"
          />
        </div>
      )}

      {/* Smooth lines (for line/area charts) */}
      {isLineType && (
        <div className="flex items-center justify-between">
          <label className="text-sm text-gray-600">Smooth Lines</label>
          <input
            type="checkbox"
            checked={style?.smooth || false}
            onChange={(e) => handleChange('smooth', e.target.checked)}
            className="rounded border-gray-300"
          />
        </div>
      )}

      {/* Stack (for bar/area charts) */}
      {(isBarType || isAreaType) && (
        <div className="flex items-center justify-between">
          <label className="text-sm text-gray-600">Stacked</label>
          <input
            type="checkbox"
            checked={style?.stack || false}
            onChange={(e) => handleChange('stack', e.target.checked)}
            className="rounded border-gray-300"
          />
        </div>
      )}

      {/* Horizontal (for bar charts) */}
      {isBarType && (
        <div className="flex items-center justify-between">
          <label className="text-sm text-gray-600">Horizontal</label>
          <input
            type="checkbox"
            checked={style?.horizontal || false}
            onChange={(e) => handleChange('horizontal', e.target.checked)}
            className="rounded border-gray-300"
          />
        </div>
      )}
    </div>
  )
}
