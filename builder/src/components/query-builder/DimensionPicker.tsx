import clsx from 'clsx'
import type { CubeDimension } from '../../api/cube'

interface DimensionPickerProps {
  dimensions: CubeDimension[]
  selected: string[]
  onChange: (dimensions: string[]) => void
}

export function DimensionPicker({ dimensions, selected, onChange }: DimensionPickerProps) {
  const toggleDimension = (dimensionName: string) => {
    if (selected.includes(dimensionName)) {
      onChange(selected.filter(d => d !== dimensionName))
    } else {
      onChange([...selected, dimensionName])
    }
  }

  const getDimensionTypeIcon = (type: string) => {
    switch (type) {
      case 'time':
        return '📅'
      case 'number':
        return '#'
      default:
        return 'T'
    }
  }

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-2">
        Dimensions
        <span className="text-gray-400 font-normal ml-1">
          ({selected.length} selected)
        </span>
      </label>
      <div className="space-y-1 max-h-40 overflow-y-auto">
        {dimensions.map((dimension) => (
          <button
            key={dimension.name}
            onClick={() => toggleDimension(dimension.name)}
            className={clsx(
              'w-full flex items-center gap-2 px-3 py-2 rounded-lg text-left text-sm transition-colors',
              selected.includes(dimension.name)
                ? 'bg-green-50 border border-green-200 text-green-700'
                : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'
            )}
          >
            <span className="w-4 h-4 flex items-center justify-center text-xs shrink-0">
              {getDimensionTypeIcon(dimension.type)}
            </span>
            <div className="flex-1 min-w-0">
              <div className="truncate">{dimension.title || dimension.shortTitle}</div>
              {dimension.description && (
                <div className="text-xs text-gray-500 truncate">
                  {dimension.description}
                </div>
              )}
            </div>
            {dimension.primaryKey && (
              <span className="text-xs text-yellow-600">🔑</span>
            )}
          </button>
        ))}
        {dimensions.length === 0 && (
          <p className="text-sm text-gray-400 text-center py-2">
            No dimensions available
          </p>
        )}
      </div>
    </div>
  )
}
