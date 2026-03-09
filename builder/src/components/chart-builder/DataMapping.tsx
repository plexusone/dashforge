import type { ChartConfig, GeometryType } from '../../types/dashboard'

interface DataMappingProps {
  geometry: GeometryType
  encodings: ChartConfig['encodings']
  onChange: (encodings: ChartConfig['encodings']) => void
}

export function DataMapping({ geometry, encodings, onChange }: DataMappingProps) {
  const isPieType = geometry === 'pie' || geometry === 'funnel'

  const handleChange = (field: string, value: string) => {
    onChange({
      ...encodings,
      [field]: value
    })
  }

  // Get available fields (in a real app, this would come from the data source)
  const availableFields = [
    { value: '', label: 'Select field...' },
    { value: 'date', label: 'Date' },
    { value: 'month', label: 'Month' },
    { value: 'category', label: 'Category' },
    { value: 'region', label: 'Region' },
    { value: 'product', label: 'Product' },
    { value: 'revenue', label: 'Revenue' },
    { value: 'quantity', label: 'Quantity' },
    { value: 'profit', label: 'Profit' },
    { value: 'cost', label: 'Cost' }
  ]

  return (
    <div className="space-y-3">
      <label className="block text-xs font-medium text-gray-600">
        Data Mapping
      </label>

      {isPieType ? (
        // Pie/Funnel: value and category
        <>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Category</label>
            <select
              value={encodings?.category || ''}
              onChange={(e) => handleChange('category', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500"
            >
              {availableFields.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Value</label>
            <select
              value={encodings?.value || ''}
              onChange={(e) => handleChange('value', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500"
            >
              {availableFields.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
        </>
      ) : (
        // Cartesian charts: x, y, color
        <>
          <div>
            <label className="block text-xs text-gray-500 mb-1">X Axis</label>
            <select
              value={encodings?.x || ''}
              onChange={(e) => handleChange('x', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500"
            >
              {availableFields.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Y Axis</label>
            <select
              value={encodings?.y || ''}
              onChange={(e) => handleChange('y', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500"
            >
              {availableFields.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Color (optional)</label>
            <select
              value={encodings?.color || ''}
              onChange={(e) => handleChange('color', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500"
            >
              {availableFields.map((f) => (
                <option key={f.value} value={f.value}>{f.label}</option>
              ))}
            </select>
          </div>
        </>
      )}
    </div>
  )
}
