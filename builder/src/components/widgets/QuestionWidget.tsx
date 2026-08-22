import { useEffect, useMemo, useState } from 'react'
import { BarChart3, LineChart } from 'lucide-react'
import {
  executeAnalyticsQuery,
  getSavedQuestion,
  type AnalyticsQueryResult,
  type SavedQuestion,
} from '../../api/dashforge'
import type { Widget } from '../../types/dashboard'

interface QuestionWidgetProps {
  widget: Widget
}

type VisualizationType = 'table' | 'bar' | 'line' | 'metric'

export function QuestionWidget({ widget }: QuestionWidgetProps) {
  const [question, setQuestion] = useState<SavedQuestion | null>(null)
  const [result, setResult] = useState<AnalyticsQueryResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const visualization = useMemo(
    () => visualizationType(widget.visualization?.type ?? question?.visualization?.type),
    [question?.visualization, widget.visualization],
  )

  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!widget.questionId) {
        setQuestion(null)
        setResult(null)
        setError(null)
        return
      }
      setLoading(true)
      setError(null)
      try {
        const nextQuestion = await getSavedQuestion(widget.questionId)
        const nextResult = await executeAnalyticsQuery({
          sourceId: nextQuestion.sourceId,
          dialect: 'grokifyql',
          query: nextQuestion.query,
          limit: 500,
        })
        if (!cancelled) {
          setQuestion(nextQuestion)
          setResult(nextResult)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load question')
          setQuestion(null)
          setResult(null)
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [widget.questionId])

  if (!widget.questionId) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-gray-50 text-sm text-gray-500">
        Select a saved question.
      </div>
    )
  }

  if (loading) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-white text-sm text-gray-500">
        Loading question...
      </div>
    )
  }

  if (error) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-red-50 p-4 text-sm text-red-700">
        {error}
      </div>
    )
  }

  if (!result) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-gray-50 text-sm text-gray-500">
        No result.
      </div>
    )
  }

  if (visualization === 'metric') {
    return <QuestionMetric result={result} />
  }
  if (visualization === 'bar' || visualization === 'line') {
    return <QuestionChart result={result} visualization={visualization} />
  }
  return <QuestionTable result={result} />
}

function QuestionTable({ result }: { result: AnalyticsQueryResult }) {
  return (
    <div className="h-full w-full overflow-auto">
      <table className="w-full text-sm">
        <thead className="sticky top-0 bg-gray-50">
          <tr>
            {result.columns.map((column) => (
              <th key={column.name} className="border-b px-3 py-2 text-left text-xs font-medium text-gray-500">
                {column.name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {result.rows.map((row, index) => (
            <tr key={index}>
              {result.columns.map((column) => (
                <td key={column.name} className="px-3 py-2 text-gray-800">
                  {String(row[column.name] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function QuestionMetric({ result }: { result: AnalyticsQueryResult }) {
  const firstColumn = result.columns[0]?.name
  const value = firstColumn ? result.rows[0]?.[firstColumn] : result.rowCount
  return (
    <div className="h-full w-full flex flex-col items-center justify-center p-4">
      <div className="text-3xl font-bold text-gray-900">{formatValue(value ?? result.rowCount)}</div>
      <div className="mt-1 text-xs text-gray-500">{firstColumn || 'rows'}</div>
    </div>
  )
}

function QuestionChart({
  result,
  visualization,
}: {
  result: AnalyticsQueryResult
  visualization: 'bar' | 'line'
}) {
  const categoryColumn = result.columns[0]?.name
  const valueColumn = result.columns.find((column) => column.name !== categoryColumn)?.name
  const points = result.rows.slice(0, 12).map((row) => ({
    label: categoryColumn ? String(row[categoryColumn] ?? '') : '',
    value: numericValue(valueColumn ? row[valueColumn] : undefined),
  }))
  const max = Math.max(...points.map((point) => point.value), 1)
  const Icon = visualization === 'bar' ? BarChart3 : LineChart
  return (
    <div className="h-full w-full p-4">
      <div className="mb-3 flex items-center gap-2 text-xs font-medium text-gray-500">
        <Icon className="h-4 w-4" />
        {valueColumn || 'value'}
      </div>
      <div className="flex h-[calc(100%-1.75rem)] items-end gap-2">
        {points.map((point, index) => (
          <div key={index} className="flex min-w-0 flex-1 flex-col items-center gap-1">
            <div
              className="w-full rounded-t bg-primary-400"
              style={{ height: `${Math.max(4, (point.value / max) * 100)}%` }}
            />
            <div className="w-full truncate text-center text-[10px] text-gray-500">{point.label}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

function visualizationType(value: unknown): VisualizationType {
  if (value === 'bar' || value === 'line' || value === 'metric') return value
  return 'table'
}

function numericValue(value: unknown): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function formatValue(value: unknown): string {
  if (typeof value === 'number') return new Intl.NumberFormat('en-US').format(value)
  return String(value ?? '')
}
