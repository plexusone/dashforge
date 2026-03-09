import { ArrowUp, ArrowDown, Minus } from 'lucide-react'
import type { Widget, MetricConfig } from '../../types/dashboard'
import clsx from 'clsx'

interface MetricWidgetProps {
  widget: Widget
}

// Sample data for preview
const sampleValue = 42350
const sampleChange = 12.5

export function MetricWidget({ widget }: MetricWidgetProps) {
  const config = widget.config as MetricConfig

  const formatValue = (value: number): string => {
    switch (config.format) {
      case 'currency':
        return new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: 0,
          maximumFractionDigits: 0
        }).format(value)

      case 'percent':
        return new Intl.NumberFormat('en-US', {
          style: 'percent',
          minimumFractionDigits: 1,
          maximumFractionDigits: 1
        }).format(value / 100)

      case 'compact':
        return new Intl.NumberFormat('en-US', {
          notation: 'compact',
          compactDisplay: 'short'
        }).format(value)

      default:
        return new Intl.NumberFormat('en-US').format(value)
    }
  }

  const getChangeIcon = () => {
    if (sampleChange > 0) return <ArrowUp className="w-4 h-4" />
    if (sampleChange < 0) return <ArrowDown className="w-4 h-4" />
    return <Minus className="w-4 h-4" />
  }

  const getChangeColor = () => {
    if (sampleChange > 0) return 'text-green-600'
    if (sampleChange < 0) return 'text-red-600'
    return 'text-gray-500'
  }

  return (
    <div className="h-full w-full flex flex-col items-center justify-center p-4">
      {/* Title */}
      {widget.title && (
        <div className="text-sm text-gray-500 mb-1">{widget.title}</div>
      )}

      {/* Main Value */}
      <div className="text-3xl font-bold text-gray-900">
        {config.prefix}
        {formatValue(sampleValue)}
        {config.suffix}
      </div>

      {/* Comparison */}
      {config.comparison?.enabled && (
        <div className={clsx(
          'flex items-center gap-1 mt-2 text-sm font-medium',
          getChangeColor()
        )}>
          {getChangeIcon()}
          <span>{Math.abs(sampleChange)}%</span>
          <span className="text-gray-400 font-normal">vs prev</span>
        </div>
      )}

      {/* Sparkline placeholder */}
      {config.sparkline?.enabled && (
        <div className="w-full h-8 mt-3 flex items-end justify-center gap-0.5">
          {[30, 45, 35, 50, 40, 55, 60, 45, 70, 65].map((h, i) => (
            <div
              key={i}
              className="w-2 bg-primary-200 rounded-t"
              style={{ height: `${h}%` }}
            />
          ))}
        </div>
      )}
    </div>
  )
}
