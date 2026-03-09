import { ReactNode } from 'react'
import { X } from 'lucide-react'
import clsx from 'clsx'

interface SidebarProps {
  title: string
  children: ReactNode
  position?: 'left' | 'right'
  width?: string
  onClose?: () => void
  className?: string
}

export function Sidebar({
  title,
  children,
  position = 'left',
  width = 'w-64',
  onClose,
  className
}: SidebarProps) {
  return (
    <aside
      className={clsx(
        'flex flex-col bg-white border-gray-200 overflow-hidden',
        position === 'left' ? 'border-r' : 'border-l',
        width,
        className
      )}
    >
      {/* Header */}
      <div className="h-12 px-4 border-b border-gray-200 flex items-center justify-between shrink-0">
        <h2 className="font-medium text-gray-900">{title}</h2>
        {onClose && (
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 rounded transition-colors"
          >
            <X className="w-4 h-4 text-gray-500" />
          </button>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {children}
      </div>
    </aside>
  )
}

interface SidebarSectionProps {
  title?: string
  children: ReactNode
  className?: string
  collapsible?: boolean
  defaultCollapsed?: boolean
}

export function SidebarSection({
  title,
  children,
  className
}: SidebarSectionProps) {
  return (
    <div className={clsx('border-b border-gray-100 last:border-b-0', className)}>
      {title && (
        <div className="px-4 py-2 bg-gray-50">
          <h3 className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            {title}
          </h3>
        </div>
      )}
      <div className="p-4">
        {children}
      </div>
    </div>
  )
}
