import { GeometryPicker } from './GeometryPicker'
import { DataMapping } from './DataMapping'
import { StyleEditor } from './StyleEditor'
import type { ChartConfig, GeometryType } from '../../types/dashboard'

interface ChartBuilderProps {
  config: ChartConfig
  onChange: (config: ChartConfig) => void
}

export function ChartBuilder({ config, onChange }: ChartBuilderProps) {
  const handleGeometryChange = (geometry: GeometryType) => {
    // Reset encodings based on geometry type
    const newEncodings = geometry === 'pie' || geometry === 'funnel'
      ? { value: config.encodings?.value || '', category: config.encodings?.category || '' }
      : { x: config.encodings?.x || '', y: config.encodings?.y || '' }

    onChange({
      ...config,
      geometry,
      encodings: newEncodings
    })
  }

  const handleEncodingsChange = (encodings: ChartConfig['encodings']) => {
    onChange({
      ...config,
      encodings
    })
  }

  const handleStyleChange = (style: ChartConfig['style']) => {
    onChange({
      ...config,
      style
    })
  }

  return (
    <div className="space-y-4">
      {/* Geometry Selection */}
      <div>
        <label className="block text-xs font-medium text-gray-600 mb-2">
          Chart Type
        </label>
        <GeometryPicker
          value={config.geometry}
          onChange={handleGeometryChange}
        />
      </div>

      {/* Data Mapping */}
      <DataMapping
        geometry={config.geometry}
        encodings={config.encodings || {}}
        onChange={handleEncodingsChange}
      />

      {/* Style Options */}
      <StyleEditor
        geometry={config.geometry}
        style={config.style || {}}
        onChange={handleStyleChange}
      />
    </div>
  )
}
