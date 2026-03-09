import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Play, RefreshCw } from 'lucide-react'
import { CubeSelector } from './CubeSelector'
import { MeasurePicker } from './MeasurePicker'
import { DimensionPicker } from './DimensionPicker'
import { FilterBuilder } from './FilterBuilder'
import { QueryPreview } from './QueryPreview'
import { fetchSchema, executeQuery, buildQuery, type QueryBuilderInput } from '../../api/cube'

interface QueryBuilderProps {
  onQueryResult?: (data: unknown[]) => void
}

export function QueryBuilder({ onQueryResult }: QueryBuilderProps) {
  const [selectedCube, setSelectedCube] = useState<string>('')
  const [measures, setMeasures] = useState<string[]>([])
  const [dimensions, setDimensions] = useState<string[]>([])
  const [filters, setFilters] = useState<QueryBuilderInput['filters']>([])
  const [isExecuting, setIsExecuting] = useState(false)
  const [queryResult, setQueryResult] = useState<unknown[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  // Fetch Cube schema
  const { data: schema, isLoading: isLoadingSchema, error: schemaError } = useQuery({
    queryKey: ['cube-schema'],
    queryFn: fetchSchema,
    staleTime: 5 * 60 * 1000 // 5 minutes
  })

  const selectedCubeSchema = schema?.cubes.find(c => c.name === selectedCube)

  const handleExecuteQuery = async () => {
    if (!measures.length && !dimensions.length) {
      setError('Please select at least one measure or dimension')
      return
    }

    setIsExecuting(true)
    setError(null)

    try {
      const query = buildQuery({
        measures,
        dimensions,
        filters
      })

      const resultSet = await executeQuery(query)
      const data = resultSet.tablePivot()
      setQueryResult(data)
      onQueryResult?.(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Query failed')
    } finally {
      setIsExecuting(false)
    }
  }

  const handleCubeChange = (cubeName: string) => {
    setSelectedCube(cubeName)
    // Clear selections when cube changes
    setMeasures([])
    setDimensions([])
    setFilters([])
    setQueryResult(null)
  }

  if (isLoadingSchema) {
    return (
      <div className="p-4 text-center text-gray-500">
        <RefreshCw className="w-5 h-5 animate-spin mx-auto mb-2" />
        <p className="text-sm">Loading schema...</p>
      </div>
    )
  }

  if (schemaError) {
    return (
      <div className="p-4 text-center">
        <p className="text-sm text-red-500 mb-2">Failed to load Cube schema</p>
        <p className="text-xs text-gray-500">
          Make sure Cube.js is running at http://localhost:4000
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Cube Selector */}
      <CubeSelector
        cubes={schema?.cubes || []}
        selectedCube={selectedCube}
        onChange={handleCubeChange}
      />

      {selectedCubeSchema && (
        <>
          {/* Measures */}
          <MeasurePicker
            measures={selectedCubeSchema.measures}
            selected={measures}
            onChange={setMeasures}
          />

          {/* Dimensions */}
          <DimensionPicker
            dimensions={selectedCubeSchema.dimensions}
            selected={dimensions}
            onChange={setDimensions}
          />

          {/* Filters */}
          <FilterBuilder
            dimensions={selectedCubeSchema.dimensions}
            measures={selectedCubeSchema.measures}
            filters={filters || []}
            onChange={setFilters}
          />

          {/* Execute Button */}
          <div className="pt-2">
            <button
              onClick={handleExecuteQuery}
              disabled={isExecuting || (!measures.length && !dimensions.length)}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isExecuting ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              <span>{isExecuting ? 'Running...' : 'Run Query'}</span>
            </button>
          </div>

          {/* Error */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {error}
            </div>
          )}

          {/* Results Preview */}
          {queryResult && (
            <QueryPreview data={queryResult} />
          )}
        </>
      )}
    </div>
  )
}
