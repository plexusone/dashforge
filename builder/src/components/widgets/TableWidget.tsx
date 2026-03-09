import { useState } from 'react'
import { ChevronUp, ChevronDown, ChevronLeft, ChevronRight } from 'lucide-react'
import type { Widget, TableConfig } from '../../types/dashboard'
import clsx from 'clsx'

interface TableWidgetProps {
  widget: Widget
}

// Sample data for preview
const sampleData = [
  { id: 1, name: 'Product A', category: 'Electronics', revenue: 12500, units: 450 },
  { id: 2, name: 'Product B', category: 'Clothing', revenue: 8300, units: 720 },
  { id: 3, name: 'Product C', category: 'Electronics', revenue: 15700, units: 320 },
  { id: 4, name: 'Product D', category: 'Home', revenue: 6200, units: 180 },
  { id: 5, name: 'Product E', category: 'Clothing', revenue: 9100, units: 560 },
  { id: 6, name: 'Product F', category: 'Electronics', revenue: 11200, units: 410 },
  { id: 7, name: 'Product G', category: 'Home', revenue: 4500, units: 290 },
  { id: 8, name: 'Product H', category: 'Clothing', revenue: 7800, units: 630 },
]

const defaultColumns: TableConfig['columns'] = [
  { field: 'name', header: 'Name' },
  { field: 'category', header: 'Category' },
  { field: 'revenue', header: 'Revenue', align: 'right' },
  { field: 'units', header: 'Units', align: 'right' }
]

export function TableWidget({ widget }: TableWidgetProps) {
  const config = widget.config as TableConfig
  const [sortField, setSortField] = useState<string | null>(null)
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const [page, setPage] = useState(1)

  const columns = config.columns?.length ? config.columns : defaultColumns
  const pageSize = config.pagination?.pageSize || 5

  // Sort data
  const sortedData = [...sampleData].sort((a, b) => {
    if (!sortField) return 0
    const aVal = (a as Record<string, unknown>)[sortField]
    const bVal = (b as Record<string, unknown>)[sortField]
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return sortOrder === 'asc' ? aVal - bVal : bVal - aVal
    }
    return sortOrder === 'asc'
      ? String(aVal).localeCompare(String(bVal))
      : String(bVal).localeCompare(String(aVal))
  })

  // Paginate data
  const totalPages = Math.ceil(sortedData.length / pageSize)
  const paginatedData = config.pagination?.enabled
    ? sortedData.slice((page - 1) * pageSize, page * pageSize)
    : sortedData

  const handleSort = (field: string) => {
    if (!config.sortable) return
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortOrder('asc')
    }
  }

  const formatCell = (value: unknown, format?: string): string => {
    if (value === null || value === undefined) return '-'
    if (typeof value === 'number') {
      if (format === 'currency') {
        return new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD'
        }).format(value)
      }
      return new Intl.NumberFormat('en-US').format(value)
    }
    return String(value)
  }

  return (
    <div className="h-full w-full flex flex-col overflow-hidden">
      {/* Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 sticky top-0">
            <tr>
              {columns.map((col) => (
                <th
                  key={col.field}
                  className={clsx(
                    'px-3 py-2 text-xs font-medium text-gray-500 uppercase tracking-wider border-b',
                    col.align === 'right' ? 'text-right' : 'text-left',
                    config.sortable && 'cursor-pointer hover:bg-gray-100'
                  )}
                  style={{ width: col.width }}
                  onClick={() => handleSort(col.field)}
                >
                  <div className="flex items-center gap-1">
                    <span>{col.header || col.field}</span>
                    {config.sortable && sortField === col.field && (
                      sortOrder === 'asc'
                        ? <ChevronUp className="w-3 h-3" />
                        : <ChevronDown className="w-3 h-3" />
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {paginatedData.map((row, rowIdx) => (
              <tr key={rowIdx} className="hover:bg-gray-50">
                {columns.map((col) => (
                  <td
                    key={col.field}
                    className={clsx(
                      'px-3 py-2 whitespace-nowrap',
                      col.align === 'right' ? 'text-right' : 'text-left'
                    )}
                  >
                    {formatCell((row as Record<string, unknown>)[col.field], col.format)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {config.pagination?.enabled && totalPages > 1 && (
        <div className="flex items-center justify-between px-3 py-2 bg-gray-50 border-t text-xs">
          <span className="text-gray-500">
            {(page - 1) * pageSize + 1}-{Math.min(page * pageSize, sortedData.length)} of {sortedData.length}
          </span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className={clsx(
                'p-1 rounded hover:bg-gray-200',
                page === 1 && 'opacity-50 cursor-not-allowed'
              )}
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="px-2">
              {page} / {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className={clsx(
                'p-1 rounded hover:bg-gray-200',
                page === totalPages && 'opacity-50 cursor-not-allowed'
              )}
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
