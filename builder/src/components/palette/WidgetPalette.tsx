import { useEffect, useCallback } from 'react'
import {
  BarChart3,
  LineChart,
  PieChart,
  Table2,
  Hash,
  Type,
  Image,
  ScatterChart,
  AreaChart
} from 'lucide-react'
import { Sidebar, SidebarSection } from '../Sidebar'
import { useDashboardStore } from '../../stores/dashboard'
import type { Widget, WidgetType, ChartConfig, GeometryType } from '../../types/dashboard'
import clsx from 'clsx'

interface WidgetTemplate {
  type: WidgetType
  label: string
  icon: React.ReactNode
  defaultConfig: unknown
  defaultSize: { w: number; h: number }
  geometry?: GeometryType
}

const widgetTemplates: WidgetTemplate[] = [
  // Charts
  {
    type: 'chart',
    label: 'Bar Chart',
    icon: <BarChart3 className="w-5 h-5" />,
    geometry: 'bar',
    defaultConfig: {
      geometry: 'bar',
      encodings: { x: '', y: '' },
      style: { showLegend: true, showLabels: false }
    } as ChartConfig,
    defaultSize: { w: 4, h: 3 }
  },
  {
    type: 'chart',
    label: 'Line Chart',
    icon: <LineChart className="w-5 h-5" />,
    geometry: 'line',
    defaultConfig: {
      geometry: 'line',
      encodings: { x: '', y: '' },
      style: { showLegend: true, smooth: false }
    } as ChartConfig,
    defaultSize: { w: 6, h: 3 }
  },
  {
    type: 'chart',
    label: 'Area Chart',
    icon: <AreaChart className="w-5 h-5" />,
    geometry: 'area',
    defaultConfig: {
      geometry: 'area',
      encodings: { x: '', y: '' },
      style: { showLegend: true, stack: true }
    } as ChartConfig,
    defaultSize: { w: 6, h: 3 }
  },
  {
    type: 'chart',
    label: 'Pie Chart',
    icon: <PieChart className="w-5 h-5" />,
    geometry: 'pie',
    defaultConfig: {
      geometry: 'pie',
      encodings: { value: '', category: '' },
      style: { showLegend: true, legendPosition: 'right' }
    } as ChartConfig,
    defaultSize: { w: 4, h: 3 }
  },
  {
    type: 'chart',
    label: 'Scatter Plot',
    icon: <ScatterChart className="w-5 h-5" />,
    geometry: 'scatter',
    defaultConfig: {
      geometry: 'scatter',
      encodings: { x: '', y: '' },
      style: { showLegend: true }
    } as ChartConfig,
    defaultSize: { w: 4, h: 3 }
  },

  // Data widgets
  {
    type: 'metric',
    label: 'Metric',
    icon: <Hash className="w-5 h-5" />,
    defaultConfig: {
      valueField: '',
      format: 'number',
      comparison: { enabled: false },
      sparkline: { enabled: false }
    },
    defaultSize: { w: 2, h: 2 }
  },
  {
    type: 'table',
    label: 'Table',
    icon: <Table2 className="w-5 h-5" />,
    defaultConfig: {
      columns: [],
      pagination: { enabled: true, pageSize: 10 },
      sortable: true,
      filterable: false
    },
    defaultSize: { w: 6, h: 4 }
  },

  // Content widgets
  {
    type: 'text',
    label: 'Text',
    icon: <Type className="w-5 h-5" />,
    defaultConfig: {
      content: 'Enter text here...',
      format: 'markdown'
    },
    defaultSize: { w: 4, h: 2 }
  },
  {
    type: 'image',
    label: 'Image',
    icon: <Image className="w-5 h-5" />,
    defaultConfig: {
      src: '',
      alt: '',
      fit: 'contain'
    },
    defaultSize: { w: 3, h: 3 }
  }
]

interface DraggableWidgetProps {
  template: WidgetTemplate
  onDragStart: (template: WidgetTemplate) => void
}

function DraggableWidget({ template, onDragStart }: DraggableWidgetProps) {
  return (
    <div
      className={clsx(
        'flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg',
        'cursor-grab active:cursor-grabbing',
        'hover:border-primary-300 hover:bg-primary-50 transition-colors'
      )}
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData('widget-type', template.type)
        e.dataTransfer.setData('widget-template', JSON.stringify(template))
        onDragStart(template)
      }}
    >
      <div className="text-gray-500">
        {template.icon}
      </div>
      <span className="text-sm font-medium text-gray-700">
        {template.label}
      </span>
    </div>
  )
}

export function WidgetPalette() {
  const { addWidget, dashboard } = useDashboardStore()

  // Find next available Y position
  const getNextY = useCallback(() => {
    if (dashboard.widgets.length === 0) return 0
    const maxY = Math.max(...dashboard.widgets.map(w => w.position.y + w.position.h))
    return maxY
  }, [dashboard.widgets])

  // Handle widget drop events from Canvas
  useEffect(() => {
    const handleWidgetDrop = (e: CustomEvent<{
      type: WidgetType
      position: { x: number; y: number; w: number; h: number }
    }>) => {
      const template = widgetTemplates.find(t => t.type === e.detail.type)
      if (!template) return

      const widget: Widget = {
        id: crypto.randomUUID(),
        type: template.type,
        title: template.label,
        position: {
          x: e.detail.position.x,
          y: e.detail.position.y,
          w: template.defaultSize.w,
          h: template.defaultSize.h,
          minW: 1,
          minH: 1
        },
        config: template.defaultConfig
      }

      addWidget(widget)
    }

    window.addEventListener('widget-drop', handleWidgetDrop as EventListener)
    return () => window.removeEventListener('widget-drop', handleWidgetDrop as EventListener)
  }, [addWidget])

  const handleDragStart = (template: WidgetTemplate) => {
    // Visual feedback during drag
    console.log('Dragging:', template.label)
  }

  const handleQuickAdd = (template: WidgetTemplate) => {
    const widget: Widget = {
      id: crypto.randomUUID(),
      type: template.type,
      title: template.label,
      position: {
        x: 0,
        y: getNextY(),
        w: template.defaultSize.w,
        h: template.defaultSize.h,
        minW: 1,
        minH: 1
      },
      config: template.defaultConfig
    }

    addWidget(widget)
  }

  const chartWidgets = widgetTemplates.filter(t => t.type === 'chart')
  const dataWidgets = widgetTemplates.filter(t => t.type === 'metric' || t.type === 'table')
  const contentWidgets = widgetTemplates.filter(t => t.type === 'text' || t.type === 'image')

  return (
    <Sidebar title="Widgets" position="left" width="w-64">
      <SidebarSection title="Charts">
        <div className="space-y-2">
          {chartWidgets.map((template, idx) => (
            <DraggableWidget
              key={`${template.type}-${template.geometry || idx}`}
              template={template}
              onDragStart={handleDragStart}
            />
          ))}
        </div>
      </SidebarSection>

      <SidebarSection title="Data">
        <div className="space-y-2">
          {dataWidgets.map((template) => (
            <DraggableWidget
              key={template.type}
              template={template}
              onDragStart={handleDragStart}
            />
          ))}
        </div>
      </SidebarSection>

      <SidebarSection title="Content">
        <div className="space-y-2">
          {contentWidgets.map((template) => (
            <DraggableWidget
              key={template.type}
              template={template}
              onDragStart={handleDragStart}
            />
          ))}
        </div>
      </SidebarSection>

      {/* Quick Add Section */}
      <SidebarSection title="Quick Add">
        <p className="text-xs text-gray-500 mb-3">
          Click to add at the bottom of the canvas
        </p>
        <div className="grid grid-cols-3 gap-2">
          {widgetTemplates.slice(0, 6).map((template, idx) => (
            <button
              key={`${template.type}-${template.geometry || idx}`}
              onClick={() => handleQuickAdd(template)}
              className="p-2 bg-gray-50 hover:bg-primary-50 border border-gray-200 hover:border-primary-300 rounded-lg transition-colors"
              title={template.label}
            >
              <div className="text-gray-500">
                {template.icon}
              </div>
            </button>
          ))}
        </div>
      </SidebarSection>
    </Sidebar>
  )
}
