/**
 * Cube.js Client Wrapper
 * Provides semantic query capabilities and schema introspection.
 */

import cubejs, { CubeApi, Query, ResultSet, Meta, Filter } from '@cubejs-client/core'

export interface CubeConfig {
  apiUrl: string
  apiToken?: string
}

let cubeApi: CubeApi | null = null

export function initCubeClient(config: CubeConfig): CubeApi {
  cubeApi = cubejs(config.apiToken || '', {
    apiUrl: config.apiUrl
  })
  return cubeApi
}

export function getCubeClient(): CubeApi {
  if (!cubeApi) {
    // Default to local Cube.js instance
    cubeApi = cubejs('', {
      apiUrl: 'http://localhost:4000/cubejs-api/v1'
    })
  }
  return cubeApi
}

// Schema types
export interface CubeMember {
  name: string
  title: string
  shortTitle: string
  type: string
  description?: string
  format?: string
  meta?: Record<string, unknown>
}

export interface CubeMeasure extends CubeMember {
  aggType?: string
  drillMembers?: string[]
}

export interface CubeDimension extends CubeMember {
  primaryKey?: boolean
}

export interface CubeSegment extends CubeMember {}

export interface CubeDefinition {
  name: string
  title: string
  description?: string
  measures: CubeMeasure[]
  dimensions: CubeDimension[]
  segments: CubeSegment[]
}

export interface CubeSchema {
  cubes: CubeDefinition[]
}

/**
 * Fetch the Cube.js schema (meta) information
 */
export async function fetchSchema(): Promise<CubeSchema> {
  const client = getCubeClient()
  const meta: Meta = await client.meta()

  // Transform meta to our schema format
  const cubes: CubeDefinition[] = Object.entries(meta.cubes || {}).map(([, cube]) => ({
    name: cube.name,
    title: cube.title || cube.name,
    description: (cube as { description?: string }).description,
    measures: cube.measures.map((m) => ({
      name: m.name,
      title: m.title || m.name,
      shortTitle: m.shortTitle || m.name,
      type: m.type,
      description: (m as { description?: string }).description,
      aggType: m.aggType
    })),
    dimensions: cube.dimensions.map((d) => ({
      name: d.name,
      title: d.title || d.name,
      shortTitle: d.shortTitle || d.name,
      type: d.type,
      description: (d as { description?: string }).description,
      primaryKey: (d as { primaryKey?: boolean }).primaryKey
    })),
    segments: cube.segments?.map((s) => ({
      name: s.name,
      title: s.title || s.name,
      shortTitle: s.shortTitle || s.name,
      type: 'segment',
      description: (s as { description?: string }).description
    })) || []
  }))

  return { cubes }
}

/**
 * Execute a Cube.js query
 */
export async function executeQuery(query: Query): Promise<ResultSet> {
  const client = getCubeClient()
  return client.load(query)
}

/**
 * Build a Cube.js query from UI selections
 */
export interface QueryBuilderInput {
  measures?: string[]
  dimensions?: string[]
  timeDimensions?: {
    dimension: string
    granularity?: 'day' | 'week' | 'month' | 'quarter' | 'year'
    dateRange?: [string, string] | string
  }[]
  filters?: {
    member: string
    operator: string
    values?: string[]
  }[]
  order?: {
    [member: string]: 'asc' | 'desc'
  }
  limit?: number
  offset?: number
}

export function buildQuery(input: QueryBuilderInput): Query {
  return {
    measures: input.measures || [],
    dimensions: input.dimensions || [],
    timeDimensions: input.timeDimensions || [],
    filters: (input.filters || []) as Filter[],
    order: input.order,
    limit: input.limit,
    offset: input.offset
  }
}

/**
 * Convert Cube.js ResultSet to chart-friendly format
 */
export interface ChartData {
  columns: string[]
  rows: Record<string, unknown>[]
  seriesNames?: string[]
}

export function resultSetToChartData(resultSet: ResultSet): ChartData {
  const columns = resultSet.tableColumns().map(c => c.key)
  const rows = resultSet.tablePivot()

  // Extract series names for multi-series charts
  const seriesNames = resultSet.seriesNames().map(s => s.title)

  return {
    columns,
    rows,
    seriesNames: seriesNames.length > 0 ? seriesNames : undefined
  }
}

/**
 * Generate natural language description of a query
 * Useful for AI context and accessibility
 */
export function describeQuery(query: Query, schema: CubeSchema): string {
  const parts: string[] = []

  // Describe measures
  if (query.measures && query.measures.length > 0) {
    const measureNames = query.measures.map(m => {
      const [cubeName, measureName] = m.split('.')
      const cube = schema.cubes.find(c => c.name === cubeName)
      const measure = cube?.measures.find(me => me.name === m)
      return measure?.title || measureName
    })
    parts.push(`Showing ${measureNames.join(', ')}`)
  }

  // Describe dimensions
  if (query.dimensions && query.dimensions.length > 0) {
    const dimNames = query.dimensions.map(d => {
      const [cubeName, dimName] = d.split('.')
      const cube = schema.cubes.find(c => c.name === cubeName)
      const dimension = cube?.dimensions.find(di => di.name === d)
      return dimension?.title || dimName
    })
    parts.push(`by ${dimNames.join(', ')}`)
  }

  // Describe time dimension
  if (query.timeDimensions && query.timeDimensions.length > 0) {
    const td = query.timeDimensions[0]
    if (td.granularity) {
      parts.push(`grouped by ${td.granularity}`)
    }
    if (td.dateRange) {
      if (typeof td.dateRange === 'string') {
        parts.push(`for ${td.dateRange}`)
      } else {
        parts.push(`from ${td.dateRange[0]} to ${td.dateRange[1]}`)
      }
    }
  }

  // Describe filters
  if (query.filters && query.filters.length > 0) {
    const filterDescs = query.filters.map(f => {
      if ('member' in f) {
        return `${f.member} ${f.operator} ${f.values?.join(', ') || ''}`
      }
      return 'complex filter'
    })
    parts.push(`where ${filterDescs.join(' and ')}`)
  }

  return parts.join(' ') || 'Empty query'
}

// Export types from @cubejs-client/core for convenience
export type { Query, ResultSet, Meta }
