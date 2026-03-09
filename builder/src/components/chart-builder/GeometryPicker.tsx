import {
  BarChart3,
  LineChart,
  PieChart,
  ScatterChart,
  AreaChart
} from 'lucide-react'
import clsx from 'clsx'
import type { GeometryType } from '../../types/dashboard'

interface GeometryPickerProps {
  value: GeometryType
  onChange: (geometry: GeometryType) => void
}

const geometryOptions: { type: GeometryType; label: string; icon: React.ReactNode }[] = [
  { type: 'bar', label: 'Bar', icon: <BarChart3 className="w-4 h-4" /> },
  { type: 'line', label: 'Line', icon: <LineChart className="w-4 h-4" /> },
  { type: 'area', label: 'Area', icon: <AreaChart className="w-4 h-4" /> },
  { type: 'pie', label: 'Pie', icon: <PieChart className="w-4 h-4" /> },
  { type: 'scatter', label: 'Scatter', icon: <ScatterChart className="w-4 h-4" /> }
]

export function GeometryPicker({ value, onChange }: GeometryPickerProps) {
  return (
    <div className="grid grid-cols-5 gap-1">
      {geometryOptions.map((option) => (
        <button
          key={option.type}
          onClick={() => onChange(option.type)}
          className={clsx(
            'flex flex-col items-center gap-1 p-2 rounded-lg border transition-colors',
            value === option.type
              ? 'border-primary-500 bg-primary-50 text-primary-700'
              : 'border-gray-200 hover:border-gray-300 text-gray-600'
          )}
          title={option.label}
        >
          {option.icon}
          <span className="text-xs">{option.label}</span>
        </button>
      ))}
    </div>
  )
}
