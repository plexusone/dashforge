interface GridOverlayProps {
  columns: number
  rowHeight: number
  gap: number
  width: number
}

export function GridOverlay({ columns, rowHeight, gap, width }: GridOverlayProps) {
  const columnWidth = (width - gap * (columns + 1)) / columns

  return (
    <div
      className="absolute inset-0 pointer-events-none opacity-30"
      style={{
        backgroundImage: `
          linear-gradient(to right, #e5e7eb 1px, transparent 1px),
          linear-gradient(to bottom, #e5e7eb 1px, transparent 1px)
        `,
        backgroundSize: `${columnWidth + gap}px ${rowHeight + gap}px`,
        backgroundPosition: `${gap}px ${gap}px`
      }}
    />
  )
}
