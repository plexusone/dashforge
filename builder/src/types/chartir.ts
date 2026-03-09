/**
 * ChartIR Types - Based on echartify specification
 * Non-polymorphic, LLM-friendly JSON format for chart specifications.
 */

export type GeometryType =
  | 'line'
  | 'bar'
  | 'pie'
  | 'scatter'
  | 'area'
  | 'radar'
  | 'funnel'
  | 'gauge'
  | 'heatmap'
  | 'treemap'
  | 'sankey'

export interface Encoding {
  field: string
  type?: 'quantitative' | 'nominal' | 'ordinal' | 'temporal'
  aggregate?: 'sum' | 'avg' | 'min' | 'max' | 'count'
  format?: string
  title?: string
}

export interface Encodings {
  // Cartesian encodings
  x?: Encoding | string
  y?: Encoding | string

  // Common encodings
  color?: Encoding | string
  size?: Encoding | string
  label?: Encoding | string
  tooltip?: (Encoding | string)[]

  // Pie/Funnel specific
  value?: Encoding | string
  category?: Encoding | string

  // Radar specific
  angle?: Encoding | string
  radius?: Encoding | string

  // Heatmap specific
  heat?: Encoding | string
}

export interface Mark {
  type: GeometryType
  style?: MarkStyle
}

export interface MarkStyle {
  // Line/Area specific
  smooth?: boolean
  areaOpacity?: number
  lineWidth?: number
  lineDash?: number[]

  // Bar specific
  barWidth?: number | string
  barGap?: string
  borderRadius?: number | number[]

  // Point specific
  symbol?: 'circle' | 'rect' | 'roundRect' | 'triangle' | 'diamond' | 'pin' | 'arrow'
  symbolSize?: number

  // General
  opacity?: number
}

export interface Scale {
  type?: 'linear' | 'log' | 'time' | 'band' | 'point'
  domain?: [number, number] | string[]
  range?: [number, number] | string[]
  nice?: boolean
  zero?: boolean
}

export interface Axis {
  show?: boolean
  title?: string
  grid?: boolean
  scale?: Scale
  labelRotate?: number
  labelFormat?: string
}

export interface Legend {
  show?: boolean
  position?: 'top' | 'bottom' | 'left' | 'right'
  orient?: 'horizontal' | 'vertical'
}

export interface Tooltip {
  show?: boolean
  trigger?: 'item' | 'axis'
  formatter?: string
}

export interface Animation {
  enabled?: boolean
  duration?: number
  easing?: string
}

export interface ChartStyle {
  // Colors
  colors?: string[]
  backgroundColor?: string

  // Dimensions
  width?: number | string
  height?: number | string
  padding?: number | [number, number, number, number]

  // Axes
  xAxis?: Axis
  yAxis?: Axis

  // Components
  legend?: Legend
  tooltip?: Tooltip
  animation?: Animation

  // Series specific
  stack?: boolean | string
  horizontal?: boolean
}

/**
 * ChartIR - The main chart intermediate representation
 * This is the non-polymorphic format that LLMs can generate reliably.
 */
export interface ChartIR {
  // Required
  mark: Mark
  encodings: Encodings

  // Optional
  data?: unknown[]
  style?: ChartStyle
  title?: string
  subtitle?: string

  // Metadata
  $schema?: string
  version?: string
}

/**
 * Creates a default ChartIR for a given geometry type
 */
export function createDefaultChartIR(geometry: GeometryType): ChartIR {
  const defaults: Record<GeometryType, Partial<ChartIR>> = {
    line: {
      mark: { type: 'line', style: { smooth: false } },
      encodings: { x: '', y: '' },
      style: { xAxis: { show: true }, yAxis: { show: true }, legend: { show: true } }
    },
    bar: {
      mark: { type: 'bar' },
      encodings: { x: '', y: '' },
      style: { xAxis: { show: true }, yAxis: { show: true }, legend: { show: true } }
    },
    pie: {
      mark: { type: 'pie' },
      encodings: { value: '', category: '' },
      style: { legend: { show: true, position: 'right' } }
    },
    scatter: {
      mark: { type: 'scatter', style: { symbol: 'circle', symbolSize: 10 } },
      encodings: { x: '', y: '' },
      style: { xAxis: { show: true }, yAxis: { show: true } }
    },
    area: {
      mark: { type: 'area', style: { areaOpacity: 0.3 } },
      encodings: { x: '', y: '' },
      style: { xAxis: { show: true }, yAxis: { show: true }, stack: true }
    },
    radar: {
      mark: { type: 'radar' },
      encodings: { angle: '', radius: '' },
      style: { legend: { show: true } }
    },
    funnel: {
      mark: { type: 'funnel' },
      encodings: { value: '', category: '' },
      style: { legend: { show: true } }
    },
    gauge: {
      mark: { type: 'gauge' },
      encodings: { value: '' },
      style: {}
    },
    heatmap: {
      mark: { type: 'heatmap' },
      encodings: { x: '', y: '', heat: '' },
      style: { xAxis: { show: true }, yAxis: { show: true } }
    },
    treemap: {
      mark: { type: 'treemap' },
      encodings: { value: '', category: '' },
      style: {}
    },
    sankey: {
      mark: { type: 'sankey' },
      encodings: {},
      style: {}
    }
  }

  return {
    mark: defaults[geometry]?.mark || { type: geometry },
    encodings: defaults[geometry]?.encodings || {},
    style: defaults[geometry]?.style || {},
    data: []
  }
}

/**
 * Validates a ChartIR object
 */
export function validateChartIR(chart: ChartIR): { valid: boolean; errors: string[] } {
  const errors: string[] = []

  if (!chart.mark) {
    errors.push('Missing required field: mark')
  } else if (!chart.mark.type) {
    errors.push('Missing required field: mark.type')
  }

  if (!chart.encodings) {
    errors.push('Missing required field: encodings')
  }

  // Geometry-specific validation
  if (chart.mark?.type) {
    const type = chart.mark.type
    if ((type === 'line' || type === 'bar' || type === 'scatter' || type === 'area') &&
        !chart.encodings?.x && !chart.encodings?.y) {
      errors.push(`${type} chart requires x and y encodings`)
    }
    if ((type === 'pie' || type === 'funnel') &&
        !chart.encodings?.value && !chart.encodings?.category) {
      errors.push(`${type} chart requires value and category encodings`)
    }
  }

  return {
    valid: errors.length === 0,
    errors
  }
}
