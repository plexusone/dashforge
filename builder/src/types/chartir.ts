/**
 * ChartIR Types - Powered by @grokify/echartify
 *
 * This module re-exports echartify types with DashForge-specific extensions
 * for backward compatibility.
 */

import {
  type Geometry,
  type Encode,
  type Mark as EchartifyMark,
  type Style,
  type ChartIR as EchartifyChartIR,
  type Dataset,
  type Axis,
  type Legend,
  type Tooltip,
  type Grid,
  chartIRSchema,
  markSchema,
  encodeSchema,
  geometrySchema,
  styleSchema,
  datasetSchema,
  axisSchema,
  legendSchema,
  tooltipSchema,
  gridSchema,
  compile,
} from '@grokify/echartify'

// Re-export types with DashForge-friendly names
export type GeometryType = Geometry
export type Encodings = Encode
export type Encoding = Encode
export type Mark = EchartifyMark
export type MarkStyle = Style
export type ChartIR = EchartifyChartIR

// Re-export for direct use
export type { Dataset, Axis, Legend, Tooltip, Grid }

// Re-export schemas and compiler
export {
  chartIRSchema,
  markSchema,
  encodeSchema,
  geometrySchema,
  styleSchema,
  datasetSchema,
  axisSchema,
  legendSchema,
  tooltipSchema,
  gridSchema,
  compile,
}

// ChartStyle is an alias for the top-level chart styling
// (combines style properties that apply to the whole chart)
export interface ChartStyle {
  // Colors
  colors?: string[]
  backgroundColor?: string

  // Dimensions
  width?: number | string
  height?: number | string
  padding?: number | [number, number, number, number]

  // Axes
  xAxis?: {
    show?: boolean
    title?: string
    grid?: boolean
    labelRotate?: number
    labelFormat?: string
  }
  yAxis?: {
    show?: boolean
    title?: string
    grid?: boolean
    labelRotate?: number
    labelFormat?: string
  }

  // Components
  legend?: {
    show?: boolean
    position?: 'top' | 'bottom' | 'left' | 'right'
    orient?: 'horizontal' | 'vertical'
  }
  tooltip?: {
    show?: boolean
    trigger?: 'item' | 'axis'
    formatter?: string
  }
  animation?: {
    enabled?: boolean
    duration?: number
    easing?: string
  }

  // Series specific
  stack?: boolean | string
  horizontal?: boolean
}

/**
 * Validates a ChartIR object using Zod schema.
 * Returns validation result with errors if invalid.
 */
export function validateChartIR(chart: unknown): { valid: boolean; errors: string[] } {
  const result = chartIRSchema.safeParse(chart)

  if (result.success) {
    return { valid: true, errors: [] }
  }

  const errors = result.error.errors.map(
    (err: { path: (string | number)[]; message: string }) =>
      `${err.path.join('.')}: ${err.message}`
  )

  return { valid: false, errors }
}

/**
 * Creates a default ChartIR for a given geometry type.
 * This is a convenience function for the builder UI.
 */
export function createDefaultChartIR(geometry: GeometryType): ChartIR {
  const defaultDataset: Dataset = {
    id: 'default',
    columns: [],
    rows: [],
  }

  const createMark = (id: string, encode: Encode, smooth?: boolean): Mark => ({
    id,
    datasetId: 'default',
    geometry,
    encode,
    ...(smooth !== undefined ? { smooth } : {}),
  })

  const defaults: Record<GeometryType, { marks: Mark[] }> = {
    line: {
      marks: [createMark('line-1', { x: '', y: '' })],
    },
    bar: {
      marks: [createMark('bar-1', { x: '', y: '' })],
    },
    pie: {
      marks: [createMark('pie-1', { value: '', name: '' })],
    },
    scatter: {
      marks: [createMark('scatter-1', { x: '', y: '' })],
    },
    area: {
      marks: [createMark('area-1', { x: '', y: '' }, true)],
    },
    radar: {
      marks: [createMark('radar-1', {})],
    },
    funnel: {
      marks: [createMark('funnel-1', { value: '', name: '' })],
    },
    gauge: {
      marks: [createMark('gauge-1', { value: '' })],
    },
    heatmap: {
      marks: [createMark('heatmap-1', { x: '', y: '', heat: '' })],
    },
    treemap: {
      marks: [createMark('treemap-1', { value: '', name: '' })],
    },
    sankey: {
      marks: [createMark('sankey-1', { source: '', target: '', value: '' })],
    },
  }

  return {
    datasets: [defaultDataset],
    marks: defaults[geometry]?.marks || [createMark(`${geometry}-1`, {})],
  }
}
