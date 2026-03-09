import { useState } from 'react'
import { RefreshCw, Play } from 'lucide-react'
import { executeQuery, type Query } from '../../api/cube'

interface DataPreviewProps {
  cubeName: string
  dimensions: string[]
  measures: string[]
}

export function DataPreview({ dimensions, measures }: DataPreviewProps) {
  const [data, setData] = useState<Record<string, unknown>[] | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchPreview = async () => {
    if (!dimensions.length && !measures.length) {
      setError('Select at least one dimension or measure')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const query: Query = {
        measures: measures.slice(0, 3),
        dimensions: dimensions.slice(0, 3),
        limit: 10
      }

      const resultSet = await executeQuery(query)
      setData(resultSet.tablePivot())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch preview')
    } finally {
      setIsLoading(false)
    }
  }

  const columns = data && data.length > 0 ? Object.keys(data[0]) : []

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-gray-50 border-b border-gray-200">
        <span className="text-xs font-medium text-gray-600">Data Preview</span>
        <button
          onClick={fetchPreview}
          disabled={isLoading}
          className="flex items-center gap-1 px-2 py-1 text-xs text-primary-600 hover:bg-primary-50 rounded"
        >
          {isLoading ? (
            <RefreshCw className="w-3 h-3 animate-spin" />
          ) : (
            <Play className="w-3 h-3" />
          )}
          Preview
        </button>
      </div>

      {/* Content */}
      <div className="max-h-40 overflow-auto">
        {error && (
          <div className="p-3 text-sm text-red-500">{error}</div>
        )}

        {!data && !error && (
          <div className="p-4 text-center text-gray-400 text-sm">
            Click Preview to load sample data
          </div>
        )}

        {data && data.length > 0 && (
          <table className="w-full text-xs">
            <thead className="bg-gray-50 sticky top-0">
              <tr>
                {columns.map((col) => (
                  <th
                    key={col}
                    className="px-2 py-1.5 text-left font-medium text-gray-600 border-b"
                  >
                    {col.split('.').pop()}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {data.map((row, idx) => (
                <tr key={idx} className="hover:bg-gray-50">
                  {columns.map((col) => (
                    <td key={col} className="px-2 py-1.5 whitespace-nowrap">
                      {formatValue(row[col])}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {data && data.length === 0 && (
          <div className="p-3 text-center text-gray-400 text-sm">
            No data found
          </div>
        )}
      </div>
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
