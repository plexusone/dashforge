import { GripVertical, Trash2, Copy } from 'lucide-react'
import clsx from 'clsx'
import { useDashboardStore } from '../../stores/dashboard'
import { ChartWidget } from '../widgets/ChartWidget'
import { MetricWidget } from '../widgets/MetricWidget'
import { TableWidget } from '../widgets/TableWidget'
import { TextWidget } from '../widgets/TextWidget'
import { ImageWidget } from '../widgets/ImageWidget'
import type { Widget } from '../../types/dashboard'

interface WidgetContainerProps {
  widget: Widget
  isSelected: boolean
  onSelect: () => void
  isEditing: boolean
}

export function WidgetContainer({
  widget,
  isSelected,
  onSelect,
  isEditing
}: WidgetContainerProps) {
  const { removeWidget, duplicateWidget } = useDashboardStore()

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirm('Delete this widget?')) {
      removeWidget(widget.id)
    }
  }

  const handleDuplicate = (e: React.MouseEvent) => {
    e.stopPropagation()
    duplicateWidget(widget.id)
  }

  const renderWidget = () => {
    switch (widget.type) {
      case 'chart':
        return <ChartWidget widget={widget} />
      case 'metric':
        return <MetricWidget widget={widget} />
      case 'table':
        return <TableWidget widget={widget} />
      case 'text':
        return <TextWidget widget={widget} />
      case 'image':
        return <ImageWidget widget={widget} />
      default:
        return (
          <div className="w-full h-full flex items-center justify-center bg-gray-50">
            <span className="text-gray-400">Unknown widget type</span>
          </div>
        )
    }
  }

  return (
    <div
      className={clsx(
        'h-full w-full bg-white rounded-lg border-2 overflow-hidden transition-all',
        isSelected
          ? 'border-primary-500 ring-2 ring-primary-200'
          : 'border-gray-200 hover:border-gray-300',
        isEditing && 'cursor-move'
      )}
      onClick={(e) => {
        e.stopPropagation()
        onSelect()
      }}
    >
      {/* Widget Header (visible in edit mode) */}
      {isEditing && (
        <div className="h-8 px-2 bg-gray-50 border-b border-gray-200 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-2">
            <div className="widget-drag-handle cursor-grab active:cursor-grabbing p-1 hover:bg-gray-200 rounded">
              <GripVertical className="w-3 h-3 text-gray-400" />
            </div>
            <span className="text-xs font-medium text-gray-600 truncate max-w-32">
              {widget.title || `${widget.type.charAt(0).toUpperCase() + widget.type.slice(1)} Widget`}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={handleDuplicate}
              className="p-1 hover:bg-gray-200 rounded transition-colors"
              title="Duplicate"
            >
              <Copy className="w-3 h-3 text-gray-400" />
            </button>
            <button
              onClick={handleDelete}
              className="p-1 hover:bg-red-100 rounded transition-colors"
              title="Delete"
            >
              <Trash2 className="w-3 h-3 text-gray-400 hover:text-red-500" />
            </button>
          </div>
        </div>
      )}

      {/* Widget Content */}
      <div className={clsx(
        'w-full overflow-hidden',
        isEditing ? 'h-[calc(100%-2rem)]' : 'h-full'
      )}>
        {renderWidget()}
      </div>
    </div>
  )
}
