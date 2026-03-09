import { useState } from 'react'
import { Table2, Code } from 'lucide-react'
import clsx from 'clsx'

interface QueryPreviewProps {
  data: unknown[]
}

export function QueryPreview({ data }: QueryPreviewProps) {
  const [viewMode, setViewMode] = useState<'table' | 'json'>('table')

  if (!data || data.length === 0) {
    return (
      <div className="p-4 bg-gray-50 rounded-lg text-center text-gray-500 text-sm">
        No results
      </div>
    )
  }

  const columns = Object.keys(data[0] as object)

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-gray-50 border-b border-gray-200">
        <span className="text-xs font-medium text-gray-600">
          {data.length} row{data.length !== 1 ? 's' : ''}
        </span>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setViewMode('table')}
            className={clsx(
              'p-1 rounded',
              viewMode === 'table' ? 'bg-white shadow' : 'hover:bg-gray-200'
            )}
          >
            <Table2 className="w-4 h-4" />
          </button>
          <button
            onClick={() => setViewMode('json')}
            className={clsx(
              'p-1 rounded',
              viewMode === 'json' ? 'bg-white shadow' : 'hover:bg-gray-200'
            )}
          >
            <Code className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Content */}
      {viewMode === 'table' ? (
        <div className="max-h-60 overflow-auto">
          <table className="w-full text-xs">
            <thead className="bg-gray-50 sticky top-0">
              <tr>
                {columns.map((col) => (
                  <th
                    key={col}
                    className="px-2 py-1.5 text-left font-medium text-gray-600 border-b"
                  >
                    {col}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {data.slice(0, 50).map((row, idx) => (
                <tr key={idx} className="hover:bg-gray-50">
                  {columns.map((col) => (
                    <td key={col} className="px-2 py-1.5 whitespace-nowrap">
                      {formatValue((row as Record<string, unknown>)[col])}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          {data.length > 50 && (
            <div className="px-2 py-1 text-center text-xs text-gray-400 bg-gray-50 border-t">
              Showing first 50 of {data.length} rows
            </div>
          )}
        </div>
      ) : (
        <div className="max-h-60 overflow-auto p-2">
          <pre className="text-xs text-gray-700 whitespace-pre-wrap">
            {JSON.stringify(data.slice(0, 10), null, 2)}
          </pre>
          {data.length > 10 && (
            <div className="text-center text-xs text-gray-400 pt-2">
              Showing first 10 of {data.length} rows
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'number') {
    return Number.isInteger(value)
      ? value.toLocaleString()
      : value.toLocaleString(undefined, { maximumFractionDigits: 2 })
  }
  return String(value)
}
