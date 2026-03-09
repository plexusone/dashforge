import { Hash } from 'lucide-react'
import clsx from 'clsx'
import type { CubeMeasure } from '../../api/cube'

interface MeasurePickerProps {
  measures: CubeMeasure[]
  selected: string[]
  onChange: (measures: string[]) => void
}

export function MeasurePicker({ measures, selected, onChange }: MeasurePickerProps) {
  const toggleMeasure = (measureName: string) => {
    if (selected.includes(measureName)) {
      onChange(selected.filter(m => m !== measureName))
    } else {
      onChange([...selected, measureName])
    }
  }

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-2">
        Measures
        <span className="text-gray-400 font-normal ml-1">
          ({selected.length} selected)
        </span>
      </label>
      <div className="space-y-1 max-h-40 overflow-y-auto">
        {measures.map((measure) => (
          <button
            key={measure.name}
            onClick={() => toggleMeasure(measure.name)}
            className={clsx(
              'w-full flex items-center gap-2 px-3 py-2 rounded-lg text-left text-sm transition-colors',
              selected.includes(measure.name)
                ? 'bg-primary-50 border border-primary-200 text-primary-700'
                : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'
            )}
          >
            <Hash className="w-3 h-3 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="truncate">{measure.title || measure.shortTitle}</div>
              {measure.description && (
                <div className="text-xs text-gray-500 truncate">
                  {measure.description}
                </div>
              )}
            </div>
            {measure.aggType && (
              <span className="text-xs text-gray-400 uppercase">
                {measure.aggType}
              </span>
            )}
          </button>
        ))}
        {measures.length === 0 && (
          <p className="text-sm text-gray-400 text-center py-2">
            No measures available
          </p>
        )}
      </div>
    </div>
  )
}
