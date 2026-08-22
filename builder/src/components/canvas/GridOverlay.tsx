export interface GridOverlayProps {
  columns: number
  rowHeight: number
  gap: number
  width: number
}

export function gridBackgroundStyle({ columns, rowHeight, gap, width }: GridOverlayProps) {
  const columnWidth = (width - gap * (columns + 1)) / columns
  return {
    backgroundImage: `
      linear-gradient(to right, rgba(148, 163, 184, 0.26) 1px, transparent 1px),
      linear-gradient(to bottom, rgba(148, 163, 184, 0.22) 1px, transparent 1px)
    `,
    backgroundSize: `${columnWidth + gap}px ${rowHeight + gap}px`,
    backgroundPosition: `${gap}px ${gap}px`,
  }
}

export function GridOverlay(props: GridOverlayProps) {
  return (
    <div
      className="absolute inset-0 pointer-events-none"
      style={gridBackgroundStyle(props)}
    />
  )
}
