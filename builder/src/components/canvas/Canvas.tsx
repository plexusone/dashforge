import { useCallback, useMemo } from 'react'
import GridLayout, { Layout as GridLayoutItem } from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { useDashboardStore } from '../../stores/dashboard'
import { WidgetContainer } from './WidgetContainer'
import { gridBackgroundStyle } from './GridOverlay'
// Widget type is inferred from dashboard store

interface CanvasProps {
  selectedWidgetId: string | null
  onSelectWidget: (widgetId: string | null) => void
}

export function Canvas({ selectedWidgetId, onSelectWidget }: CanvasProps) {
  const { dashboard, moveWidget, isEditing } = useDashboardStore()
  const { layout, widgets } = dashboard

  const columns = layout.columns || 12
  const rowHeight = layout.rowHeight || 80
  const gap = layout.gap || 8

  // Calculate canvas width (assuming 1200px max width, with padding)
  const canvasWidth = 1200

  // Convert widgets to react-grid-layout format
  const gridLayout: GridLayoutItem[] = useMemo(
    () =>
      widgets.map((widget) => ({
        i: widget.id,
        x: widget.position.x,
        y: widget.position.y,
        w: widget.position.w,
        h: widget.position.h,
        minW: widget.position.minW || 1,
        minH: widget.position.minH || 1,
        maxW: widget.position.maxW,
        maxH: widget.position.maxH,
        static: !isEditing,
      })),
    [widgets, isEditing],
  )

  // Handle layout changes from drag/resize
  const handleLayoutChange = useCallback(
    (newLayout: GridLayoutItem[]) => {
      newLayout.forEach((item) => {
        const widget = widgets.find((w) => w.id === item.i)
        if (widget) {
          const positionChanged =
            widget.position.x !== item.x ||
            widget.position.y !== item.y ||
            widget.position.w !== item.w ||
            widget.position.h !== item.h

          if (positionChanged) {
            moveWidget(item.i, {
              x: item.x,
              y: item.y,
              w: item.w,
              h: item.h,
            })
          }
        }
      })
    },
    [widgets, moveWidget],
  )

  // Handle click on canvas background (deselect)
  const handleCanvasClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === e.currentTarget) {
        onSelectWidget(null)
      }
    },
    [onSelectWidget],
  )

  // Handle drop from palette
  const handleDrop = useCallback(
    (_layout: GridLayoutItem[], item: GridLayoutItem, e: DragEvent) => {
      const widgetType = e.dataTransfer?.getData('widget-type')
      const widgetTemplate = e.dataTransfer?.getData('widget-template')
      if (widgetType) {
        // The WidgetPalette will handle creating the actual widget
        // This just provides the drop position
        const event = new CustomEvent('widget-drop', {
          detail: {
            type: widgetType,
            templateLabel: widgetTemplate ? JSON.parse(widgetTemplate).label : undefined,
            position: { x: item.x, y: item.y, w: item.w, h: item.h },
          },
        })
        window.dispatchEvent(event)
      }
    },
    [],
  )

  return (
    <div className="relative min-h-full" onClick={handleCanvasClick}>
      {/* Main grid layout */}
      <div
        className="relative bg-white rounded-lg shadow-sm border border-gray-200 mx-auto overflow-hidden"
        style={{
          width: canvasWidth,
          ...(isEditing ? gridBackgroundStyle({ columns, rowHeight, gap, width: canvasWidth }) : {}),
        }}
      >
        {widgets.length === 0 && isEditing ? (
          <div className="flex items-center justify-center h-96 text-gray-500">
            <div className="text-center rounded-lg border border-gray-200 bg-white/85 px-8 py-6 shadow-sm backdrop-blur-sm">
              <p className="text-lg font-medium text-gray-700 mb-2">No widgets yet</p>
              <p className="text-sm text-gray-500">Drag widgets from the palette to get started</p>
            </div>
          </div>
        ) : (
          <GridLayout
            className="layout"
            layout={gridLayout}
            cols={columns}
            rowHeight={rowHeight}
            width={canvasWidth}
            margin={[gap, gap]}
            containerPadding={[layout.padding || 16, layout.padding || 16]}
            onLayoutChange={handleLayoutChange}
            onDrop={handleDrop}
            isDroppable={isEditing}
            isDraggable={isEditing}
            isResizable={isEditing}
            compactType={null}
            preventCollision={false}
            useCSSTransforms={true}
            draggableHandle=".widget-drag-handle"
          >
            {widgets.map((widget) => (
              <div key={widget.id}>
                <WidgetContainer
                  widget={widget}
                  isSelected={selectedWidgetId === widget.id}
                  onSelect={() => onSelectWidget(widget.id)}
                  isEditing={isEditing}
                />
              </div>
            ))}
          </GridLayout>
        )}
      </div>
    </div>
  )
}
