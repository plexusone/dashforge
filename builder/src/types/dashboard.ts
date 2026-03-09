/**
 * Dashboard IR Types - TypeScript mirror of Go types in dashboardir/
 * These types match the JSON format used by the Dashforge backend.
 */

// Layout types
export type LayoutType = 'grid' | 'flex' | 'free'

export interface Layout {
  type: LayoutType
  columns?: number
  rowHeight?: number
  gap?: number
  padding?: number
}

// Theme types
export interface Theme {
  name?: string
  colors?: {
    primary?: string
    secondary?: string
    background?: string
    surface?: string
    text?: string
    border?: string
  }
  typography?: {
    fontFamily?: string
    fontSize?: number
  }
}

// Variable types
export type VariableType = 'select' | 'text' | 'date' | 'daterange'

export interface VariableOption {
  label: string
  value: string
}

export interface Variable {
  id: string
  name: string
  label?: string
  type: VariableType
  defaultValue?: string
  options?: VariableOption[]
  datasourceId?: string
  valueField?: string
  labelField?: string
}

// DataSource types
export type DataSourceType = 'url' | 'inline' | 'postgres' | 'mysql' | 'derived' | 'cube'

export interface ConnectionConfig {
  host?: string
  port?: number
  database?: string
  username?: string
  password?: string
  sslMode?: string
}

export interface RefreshConfig {
  enabled: boolean
  intervalSeconds?: number
}

export interface CacheConfig {
  enabled: boolean
  ttlSeconds?: number
}

export interface DataSource {
  id: string
  type: DataSourceType
  url?: string
  method?: string
  headers?: Record<string, string>
  body?: string
  data?: unknown
  connection?: ConnectionConfig
  query?: string
  parentId?: string
  transforms?: Transform[]
  refresh?: RefreshConfig
  cache?: CacheConfig
}

// Transform types
export type TransformType = 'extract' | 'filter' | 'aggregate' | 'sort' | 'limit' | 'select' | 'rename' | 'compute'

export interface TransformExtractConfig {
  path: string
}

export interface TransformFilterConfig {
  field: string
  operator: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'startsWith' | 'endsWith' | 'in' | 'notIn'
  value: unknown
}

export interface TransformAggregateConfig {
  groupBy: string[]
  aggregations: {
    field: string
    function: 'sum' | 'avg' | 'min' | 'max' | 'count' | 'first' | 'last'
    as: string
  }[]
}

export interface TransformSortConfig {
  field: string
  order: 'asc' | 'desc'
}

export interface TransformLimitConfig {
  limit: number
  offset?: number
}

export interface TransformSelectConfig {
  fields: string[]
}

export interface TransformRenameConfig {
  mapping: Record<string, string>
}

export interface TransformComputeConfig {
  field: string
  expression: string
}

export interface Transform {
  type: TransformType
  config: TransformExtractConfig | TransformFilterConfig | TransformAggregateConfig |
         TransformSortConfig | TransformLimitConfig | TransformSelectConfig |
         TransformRenameConfig | TransformComputeConfig
}

// Widget types
export type WidgetType = 'chart' | 'table' | 'metric' | 'text' | 'image'

export interface Position {
  x: number
  y: number
  w: number
  h: number
  minW?: number
  minH?: number
  maxW?: number
  maxH?: number
}

export interface DrillDown {
  enabled: boolean
  targetDashboardId?: string
  parameters?: Record<string, string>
}

export interface Widget {
  id: string
  type: WidgetType
  title?: string
  description?: string
  position: Position
  datasourceId?: string
  transforms?: Transform[]
  config: unknown // Type-specific config (ChartConfig, TableConfig, etc.)
  drillDown?: DrillDown
}

// Chart config types (from echartify ChartIR)
export type GeometryType = 'line' | 'bar' | 'pie' | 'scatter' | 'area' | 'radar' | 'funnel' | 'gauge' | 'heatmap' | 'treemap' | 'sankey'

export interface ChartConfig {
  geometry: GeometryType
  encodings: {
    x?: string
    y?: string
    color?: string
    size?: string
    label?: string
    value?: string      // For pie/funnel
    category?: string   // For pie/funnel
  }
  style?: {
    colors?: string[]
    showLegend?: boolean
    legendPosition?: 'top' | 'bottom' | 'left' | 'right'
    showLabels?: boolean
    smooth?: boolean    // For line/area
    stack?: boolean     // For bar/area
    horizontal?: boolean // For bar
  }
}

export interface TableConfig {
  columns: {
    field: string
    header?: string
    width?: number
    align?: 'left' | 'center' | 'right'
    format?: string
  }[]
  pagination?: {
    enabled: boolean
    pageSize?: number
  }
  sortable?: boolean
  filterable?: boolean
}

export interface MetricConfig {
  valueField: string
  format?: string
  prefix?: string
  suffix?: string
  comparison?: {
    enabled: boolean
    field?: string
    format?: string
  }
  sparkline?: {
    enabled: boolean
    field?: string
    type?: 'line' | 'bar' | 'area'
  }
}

export interface TextConfig {
  content: string
  format?: 'plain' | 'markdown' | 'html'
}

export interface ImageConfig {
  src: string
  alt?: string
  fit?: 'contain' | 'cover' | 'fill' | 'none'
}

// Dashboard type
export interface Dashboard {
  id: string
  title: string
  description?: string
  version?: string
  layout: Layout
  theme?: Theme
  variables?: Variable[]
  dataSources: DataSource[]
  widgets: Widget[]
  metadata?: Record<string, unknown>
  createdAt?: string
  updatedAt?: string
}
